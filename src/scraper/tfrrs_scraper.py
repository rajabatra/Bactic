from bs4 import BeautifulSoup as bs
from sqlalchemy import create_engine
from sqlalchemy.orm import Session
from enum import Enum
from typing import List, Tuple
import re, os, datetime, requests, random, asyncio, logging, time
import orm
import pandas as pd

field_events = {orm.EventType.high_jump, orm.EventType.vault, orm.EventType.long_jump, orm.EventType.triple_jump, orm.EventType.shot, orm.EventType.discus, orm.EventType.hammer, orm.EventType.jav, orm.EventType._4x100, orm.EventType._4x400, orm.EventType.dec, orm.EventType.hept}

lock = asyncio.Lock()

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

def parse_time(time: str) -> float:
    if time in {'DNF', 'DQ', 'FS', 'DNS', 'NT'}:
        return None
    try:
        t = datetime.datetime.strptime(time, '%M:%S.%f')
        return datetime.timedelta(minutes=t.minute, seconds=t.second, microseconds=t.microsecond).total_seconds()
    except ValueError:
        try:
            t = datetime.datetime.strptime(time, '%S.%f')
            return datetime.timedelta(seconds=t.second, microseconds=t.microsecond).total_seconds()
        except ValueError:
            raise ValueError(f'Could not parse the time {time} into one of the two acceptable formats %M:%S.%f or %S.%f')

def bucket_event(event_title: str) -> orm.EventType:
    """Convert an event title into the event enum"""
    # This may be the worst code I have ever written. I don't know any better way of doing this, but if there is, please inform me

    # strip to event
    event_title = re.sub("Men's|Women's", "", event_title).strip()
    distance = re.search(r'^\d{2,}\s', event_title)
    relay = re.search(r'4 x (\d+)m*', event_title)
    field = re.search(r'High|Long|Triple|Vault|Shot|Discus|Hammer|Javelin|Decathlon|Heptathlon', event_title)
    if distance:
        distance = int(distance.group(0))
        if distance == 100:
            following = re.search(r'^\d+\s([A-Za-z]+)', event_title).group(1)
            if following == 'Meters':
                return orm.EventType._100m
            elif following == 'Hurdles':
                return orm.EventType._100h
            else:
                raise ValueError(f'Unable to parse 100 event title {event_title}')
        elif distance == 110:
            return orm.EventType._110h
        elif distance == 200:
            return orm.EventType._200m
        elif distance == 400:
            following = re.search(r'^\d+\s([A-Za-z]+)', event_title).group(1)
            if following == 'Meters':
                return orm.EventType._400m
            elif following == 'Hurdles':
                return orm.EventType._400h
            else:
                raise ValueError(f'Unable to parse 400 event title {event_title}')
            
        elif distance == 800:
            return orm.EventType._800m
        elif distance == 1500:
            return orm.EventType._1500m
        elif distance == 5000:
            return orm.EventType._5000m
        elif distance == 3000:
            following = re.search(r'^\d+\s([A-Za-z]+)', event_title).group(1)
            if following == 'Meters':
                return orm.EventType._3000m
            elif following == 'Steeplechase':
                return orm.EventType._3000s
            else:
                raise ValueError(f'Unable to parse 3000 event title {event_title}')
        elif distance == 10000:
            return orm.EventType._10000m
        else:
            raise ValueError(f'Parsed a number {distance} that was not in the list of expected distances')
    elif relay:
        distance = int(relay.group(1))
        if distance == 100:
            return orm.EventType._4x100
        elif distance == 400:
            return orm.EventType._4x400
        else:
            raise ValueError(f'Parsed a number {distance} that was not in the list of expected relay distances')
    elif field:
        event = field.group(0)
        if event == 'Hammer':
            return orm.EventType.hammer
        elif event == 'High':
            return orm.EventType.high_jump
        elif event == 'Long':
            return orm.EventType.long_jump
        elif event == 'Triple':
            return orm.EventType.triple_jump
        elif event == 'Vault':
            return orm.EventType.vault
        elif event == 'Discus':
            return orm.EventType.discus
        elif event == 'Javelin':
            return orm.EventType.jav
        elif event == 'Shot':
            return orm.EventType.shot
        elif event == 'Decathlon':
            return orm.EventType.dec
        elif event == 'Heptathlon':
            return orm.EventType.hept
        else:
            raise ValueError(f'The event title {event} could not be parsed as a field event')
    elif event_title == '10,000 Meters':
        return orm.EventType._10000m
    else:
        raise ValueError(f'The event title {event_title} did not contain a numerical prefix')

def scrape_meet_page(root: bs) -> pd.DataFrame:
    events = root.find_all('div', {'class': 'row'})
    column_labels = ['EVENT', 'PL']
    all_events = pd.DataFrame(columns=column_labels).set_index(['EVENT', 'PL'])

    for ev in events[1:]:
        event_title = ev.h3
        if not event_title:
            # this happens when we hit heat-wise tables
            continue

        event_title = event_title.get_text().strip()

        # setup all head data for event
        try:
            event_type = bucket_event(event_title)
        except ValueError as e:
            logging.error('There was an error bucketing this event, probably due to a minor change in the event title. Check the event bucketing function against %s to see why this happened.', event_title)
            continue
        # TODO: need to parse field events at some point
        if event_type in field_events:
            continue
        head = list(map(str.strip, list(map(bs.get_text, ev.thead.find_all('th')))))
        head.append('EVENT')
        rows = ev.find_all('tr')[1:]
        data = []
        for row in rows:
            cols = row.find_all('td')
            cols_text = list(map(str.strip, map(bs.get_text, cols)))

            if 'NAME' in head:
                i = head.index('NAME')
                name = getattr(cols[i], 'a', None)
                if name:
                    cols_text[i] = re.search(r'\/(\d+)[\/\.]', name['href']).group(1)
                else:
                    cols_text[i] = None

            if 'TEAM' in head:
                i = head.index('TEAM')
                team = getattr(cols[i], 'a', None)
                if team:
                    try:
                        cols_text[i] = team['href']
                    except AttributeError:
                        print(team['href'])
                else:
                    cols_text[i] = None

            cols_text.append(event_type)
            data.append(cols_text)
        event_df = pd.DataFrame(data=data, columns=head)
        all_events = pd.concat([all_events, event_df])
    
    return all_events
    
def scrape_meet(meet_url):
    """Scrape an entire meet given the root meet page"""

    # pull both meet roots
    men_url = re.sub(r'\/(\d+)\/', r'/\1/m/', meet_url)
    men_root = get_bs(men_url)
    # slight random delay between access so we don't do this simultaneously
    women_url = re.sub(r'\/m\/', r'/f/', men_url)
    time.sleep(random.random()*10)
    women_root = get_bs(women_url)

    men_results = scrape_meet_page(men_root)
    men_results['SEX'] = orm.Sex.MALE
    women_results = scrape_meet_page(women_root)
    women_results['SEX'] = orm.Sex.FEMALE

    results = pd.concat([men_results, women_results])

    # parse columns
    if 'TIME' in results.columns:
        results['TIME'] = results['TIME'].apply(parse_time)
    results = results[results['PL'] != '']
    results['PL'] = pd.to_numeric(results['PL'], downcast='integer')
    results['NAME'] = pd.to_numeric(results['NAME'], downcast='integer')

    return results.set_index(['NAME', 'EVENT'])

def check_school(school_url: str, session: Session) -> int:
    """Provided a url to a school, assert its presence in the database or scrape its information. Return the key once finished."""
    body = get_bs(school_url)
    school_name = body.find('h3', {'id': 'team-name'}).get_text().strip()
    with lock:
        school = session.query(orm.School).filter(orm.School.name == school_name).one_or_none()
        if school:
            return school.id

        divisions = body.find('span', {'class': 'panel-heading-normal-text'})
        division = None
        if divisions:
            for d in divisions:
                division = parse_division(d.get_text())
                if division:
                    break

        # TODO: handle conference scraping, right now we have no search logic for this
        session.add(orm.School(division, school_name))
        session.flush()
        school = session.query(orm.School).filter(orm.School.name == school_name).one_or_none()
    return school.id

async def delay_scrape_athlete_and_school(athlete_id: int, sex: orm.Sex, delay: float, session: Session):
    await asyncio.sleep(delay)

    athlete_root = get_bs('https://www.tfrrs.org/athletes/' + str(athlete_id))

    name = athlete_root.find('h3', {'class': 'panel-title large-title'}).get_text()
    year = re.findall('\([A-Z]{2}-(\d)\)', name)[0]
    name = re.findall('^([A-Z\s]+)\n', name)[0]
    school_url = athlete_root.find_all('a', {'class': 'underline-hover-white pl-0 panel-actions'})[1]['href']
    school_id = check_school(school_url, session)
    ath = orm.Athlete(name=name, year=year, school_id=school_id, sex=sex)
    ath.id = athlete_id
    with lock:
        session.add(ath)
        session.flush()

async def delay_scrape(url: str, session: Session, delay: float, deadline: float):
    """Perform all scrape scheduling"""
    await asyncio.sleep(delay)

    results = scrape_meet(url)

    with asyncio.TaskGroup() as to_scrape:
        for i in results.index:
            sex = results[i]['SEX']
            athlete_id = i[0]
            with lock:
                ath = session.get(orm.Athlete, athlete_id)
            if not ath:
                delay_scrape = random.random()*(deadline - delay)
                logging.info(f'Athlete with id {athlete_id} not found. Delaying scrape in {delay_scrape:.2f} seconds')
                to_scrape(delay_scrape_athlete_and_school(athlete_id, sex, delay_scrape, session))

    with lock:
        results.to_sql('result', session.connection)
    
    try:
        session.add(results)
    except Exception as e:
        logging.error('Failed to add results table')


        




async def scrape_root(deadline, session):
    # root tfrrs call
    root = get_bs('https://www.tfrrs.org/results.rss')

    with asyncio.TaskGroup() as pending_scrapes:
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

            delay = random.uniform(0, DAY/24) # start all scraping tasks within an hour of the root scrape
            pending_scrapes.create(delay_scrape(url, session, delay))
            logging.info('Starting the scrape of %s in %s seconds', meet_title, delay)
            session.add(orm.Meet(meet_title, meet_date))


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

    # main loop
    logging.info('Starting main scrape loop')

    while True:
        try:
            now = time.time()
            asyncio.run(scrape_root(now + DAY, session))
        except Exception as e:
            logging.info('Exiting daily scrape loop and shutting down')
            break
    
    session.close()