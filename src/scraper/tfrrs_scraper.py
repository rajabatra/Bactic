from bs4 import BeautifulSoup as bs
from sqlalchemy import create_engine
from sqlalchemy.orm import Session
from enum import Enum
from typing import List, Tuple
import re, os, datetime, requests, random, asyncio, logging, time
import orm

# scrape all of the meets occurring in the past day from tfrrs.org/results.rss




# duration of one day in seconds

# enum of all event types

class Event(Enum):
    _5000m = 0


# we have some logic here that can be the general framework for our scraping strategy

# here we need to determine whether we go by team or by meet

# team advantage: we can get entire season performances at one time. With this, we can scrape on demand and will have to worry less about being detected. This seems like it might be a good method of bootstrapping the model.

# performance advantage: we get cross-sectional data of the field through time. The scraping rate doesn't need to be prohibitively fast. However, we will need some way of delineating between athletes with the same names. This seems like it would be the best method in the long-run, however because it gives a framework that can run and update periodically without user supervision.

# 1. start at latest results page and loop through all of the dates that are the current date
# For this, it will probably be best to use the rss, as this gives a full list of most recent events. Put the meet name, dates, and id in the database

def get_bs(url):
    page = requests.get(url, headers={'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64; rv:10.0) Gecko/20100101 Firefox/10.0'})
    return bs(page.content, features='lxml')

def parse_date(date_str):
    date_parsed = re.findall('([a-zA-Z]+)\s*(\d+)(\s*-\s*\d+)?,\s*(\d{4})', date_str)
    if len(date_parsed) == 0:
        raise ValueError(f'The meet date string {date_str} cannot be parsed into the regex')
    elif len(date_parsed) == 1:
        date_parsed = date_parsed[0]
        yr = int(date_parsed[3])
        mon = date_parsed[0]
        date = int(date_parsed[1]) 
        return datetime.datetime.strptime(f'{yr:04}-{mon}-{date:02}', '%Y-%B-%d').date()
    
def parse_division(region_str: str) -> orm.Division:
    if re.search('DIII', region_str):
        return orm.Division.NCAADIII
    elif re.search('DII', region_str):
        return orm.Division.NCAADII
    elif re.search('DI', region_str):
        return orm.Division.NCAADI
    elif re.search('NAIA', region_str):
        return orm.Division.NAIA
    else:
        return None

def parse_time(time: str) -> datetime.timedelta:
    try:
        t = datetime.datetime.strptime(time, '%M:%S.%f')
        return datetime.timedelta(minutes=t.minute, seconds=t.second, microseconds=t.microsecond)
    except ValueError:
        try:
            t = datetime.datetime.strptime(time, '%S.%f')
            return datetime.timedelta(seconds=t.second, microseconds=t.microsecond)
        except ValueError:
            raise ValueError(f'Could not parse the time {time} into one of the two acceptable formats %M:%S.%f or %S.%f')


def check_athlete_and_insert(athlete_id_url: str, sex: orm.Sex, session: Session, result: orm.Result, scheduler: sched.scheduler) -> int:
    # first check if athlete is in the database, if not then schedule the athlete pull with the time insertion
    athlete_id = re.match('https://www.tfrrs.org/athletes/(\d+)', athlete_id_url).group(1)
    athlete = session.get(orm.Athlete, athlete_id)
    if not athlete:
        scheduled_time = random.uniform(next_day - time.monotonic())
        
        scheduler.enter(scheduled_time, 1, populate_athlete)

    result.athlete_id = athlete.id
    session.add(orm.Result, result)

    # then check if associated 
def populate_athlete(athlete_url: str, sex: orm.Sex, result: orm.Result, session: Session):
    body = get_bs(athlete_url)
    name = body.find('h3', {'class': 'panel-title large-title'}).get_text()
    year = re.findall('\([A-Z]{2}-(\d)\)', name)[0]
    name = re.findall('^([A-Z\s]+)\n', name)[0]
    school_name = body.find_all('h3', {'class': 'panel-title'})[1].get_text().strip()
    print(school_name)
    school_url = body.find_all('a', {'class': 'underline-hover-white pl-0 panel-actions'})[1]['href']
    time.sleep(random.uniform(10))

    school_id = check_school(school_url, session)

    session.add(orm.Athlete, orm.Athlete())

def check_athlete(id_url: str, sex: orm.Sex, session: Session) -> int:
    """Check the url for the athlete's name in the database. If not found, create a child to scrape the athlete's relevant information from the url. Return the key once finished."""
    id = re.match('https://www.tfrrs.org/athletes/(\d+)',id_url).group(1)
    if not session.get(orm.Athlete, id):
        body = get_bs(id_url)
        name = body.find('h3', {'class': 'panel-title large-title'}).get_text()
        year = re.findall('\([A-Z]{2}-(\d)\)', name)[0]
        name = re.findall('^([A-Z\s]+)\n', name)[0]
        school_url = body.find_all('a', {'class': 'underline-hover-white pl-0 panel-actions'})[1]['href']
        school_id = check_school(school_url, session)
        ath = orm.Athlete(name=name, year=year, school_id=school_id, sex=sex)
        session.add(ath)
        session.flush()
    return id


def check_school(school_url: str, session: Session) -> int:
    """Provided a url to a school, assert its presence in the database or scrape its information. Return the key once finished."""
    body = get_bs(school_url)
    school_name = body.find('h3', {'id': 'team-name'}).contents
    school = session.get(orm.School, {'name': school_name})
    if school:
        return school.id

    divisions = body.find('span', {'class': 'panel-heading-normal-text'})
    division = None
    if divisions:
        for d in divisions:
            division = parse_division(d.content)
            if division:
                break

    # TODO: handle conference scraping, right now we have no search logic for this
    session.add(orm.School(division, school_name))
    session.flush()
    school = session.get(orm.School, {'name': school_name})
    return school.id
    

def parse_event_table(event_type: Event, sex: orm.Sex, date: datetime.date, body, session: Session):
    """Parse the event of a table in meet results"""
    if event_type == Event._5000m:
        for row in body.find_all('tr'):
            cells = row.find_all('td')
            pl = int(cells[0].get_text())
            time = parse_time(cells[4].get_text().strip())
            ath_url = row.a['href']
            athlete_id = check_athlete(ath_url, sex, session)
            session.add(orm.Result(athlete_id, Event._5000m, pl, date, time))
    else:
        # TODO: message that we have not yet implemented this event
        pass

def bucket_event(event_title: str) -> Event:
    """Convert an event title into the event enum"""
    if re.search('5000 Meters', event_title):
        return Event._5000m
    else:
        raise ValueError(f'The event title {event_title} could not be sorted into an event type')
    

def parse_event(body, sex: orm.Sex, date: datetime.date, session: Session) -> Tuple[Event, list]:
    """Parse the event, create corresponding ORM objects and and update the database"""
    name = body.h3.get_text()
    table = body.tbody
    event_type = bucket_event(name)

    return event_type, parse_event_table(event_type, sex, date, table, session)
    
            
    
async def scrape_meet(body, session: Session):
    """Scrape an entire meet given the root meet page"""

    events = body.find_all('div', {'class': 'row'})
    meet_results = {}
    
    for ev in events[1:]:
        ev_type, results = parse_event(ev, session)


    async with lock:
        # add all future scrapes to the queue
        pass 

async def scrape_root(deadline, session):
    # root tfrrs call
    root = get_bs('https://www.tfrrs.org/results.rss')

    pending_scrapes = asyncio.TaskGroup()

    for meet in root.rss.channel.find_all('item', recursive=False):
        meet_title = meet.title.string
        try:
            meet_date = parse_date(meet.description.string)
        except ValueError as v:
            logging.error('Meet date parsing error, not inserting into database: %s', v)
            continue
        if not meet_title:
            logging.error(f'Meet title could not be found for {meet}. Not inserting into database')
            continue
        
        try:
            url = meet.link.string
            meet_id = re.match('https://www.tfrrs.org/results/(\d+)', url).group(1)
        except AttributeError as e:
            logging.error('Unable to parse meet id from url %s due to error %s. Not pulling meet data.', url, e)
            continue

        delay = random.uniform(DAY/24) # start all scraping tasks within an hour of the root scrape
        pending_scrapes.enter(delay, 1, scrape_meet, url)
        pending_scrapes.create
        session.add(orm.Meet(meet_title, meet_date))
        # cur.execute("INSERT INTO TABLE meets(id, name, date) VALUES(%s, %s)", (id, title, date))


if __name__ == '__main__':

    n_requests = int(1e3)
    DAY = 20*60*60


    # initialize connection
    dbname = os.getenv("POSTGRES_DB")
    dbuser = os.getenv("POSTGRES_USER")
    dbpass = os.getenv("POSTGRES_PASSWORD")
    dbhost = os.getenv("POSTGRES_HOST")

    logging.basicConfig(level=logging.INFO)    

    try:
        engine = create_engine(f'postgresql+psycopg2://{dbuser}:{dbpass}@{dbhost}/{dbname}')
        session = Session(engine)
    except Exception as e:
        logging.fatal('Could not create postgres connection or session for %s', e)
    logging.info('Created session with host %s in database %s', dbhost, dbname)

    pending_scrapes = asyncio.TaskGroup()
    lock = asyncio.Lock()


    # main loop
    logging.info('Starting main scrape loop')

    while True:
        try:
            now = time.time()
            asyncio.run(scrape_root(now + DAY))
            time.sleep(DAY)
        except Exception as e:
            logging.info('Exiting daily scrape loop and shutting down')
            break
    
    session.close()


# 2. Go to the men's and women's events compiled pages
# 3. This is where I assume the logic gets a bit complicated: parse the event name at the top of each table, then loop through each of the entries and record names, There are heat tables under each of these, which I assume can be recorded to give more granular results
# If there are links associated with the athletes and schools, these should be used to populate the athlete id and school id fields
# when there is a school id that has not previously been seen, go to the school page and record its information. Wait, no maybe we should maintain a queue of pages to visit
# there are certain event types that I think will be challenging to parse. I think we should leave these for 
# we will need to write a text processor for the event names though. This shouldn't be too hard