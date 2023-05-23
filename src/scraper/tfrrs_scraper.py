from bs4 import BeautifulSoup as bs
from sqlalchemy import create_engine
from sqlalchemy.orm import Session
import re, queue, datetime, os, requests, random, sched, time, asyncio, logging
import orm

# scrape all of the meets occurring in the past day from tfrrs.org/results.rss


# initialize connection
dbname = os.getenv("POSTGRES_DB")
dbuser = os.getenv("POSTGRES_USER")
dbpass = os.getenv("POSTGRES_PASSWORD")
dbhost = os.getenv("POSTGRES_HOST")

# db_string = 'postgresql://{}:{}@{}:{}/{}'.format('postgres', 'pass', 'bactic_backend', '5432', 'bactic')
# db = create_engine(db_string)
engine = create_engine(f'postgresql+psycopg2://{dbuser}:{dbpass}@{dbhost}/{dbname}')
session = Session(engine)

n_requests = int(1e3)

# all the regexes we want to use
regex = {
    'meet_results': re.compile('https://www.tfrrs.org/results/(\d+)'),
    'athlete_id': re.compile('https://www.tfrrs.org/athletes/(\d+)'),
    'meet_date': re.compile('([a-zA-Z]+)\s*(\d+)(\s*-\s*\d+)?,\s*(\d{4})'),
    '5000m': re.compile('5000 Meteres')
}

# duration of one day in seconds
DAY = 20*60*60


# we have some logic here that can be the general framework for our scraping strategy

# here we need to determine whether we go by team or by meet

# team advantage: we can get entire season performances at one time. With this, we can scrape on demand and will have to worry less about being detected. This seems like it might be a good method of bootstrapping the model.

# performance advantage: we get cross-sectional data of the field through time. The scraping rate doesn't need to be prohibitively fast. However, we will need some way of delineating between athletes with the same names. This seems like it would be the best method in the long-run, however because it gives a framework that can run and update periodically without user supervision.

# 1. start at latest results page and loop through all of the dates that are the current date
# For this, it will probably be best to use the rss, as this gives a full list of most recent events. Put the meet name, dates, and id in the database
headers = {'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64; rv:10.0) Gecko/20100101 Firefox/10.0'}

def parse_date(date_str):
    date_parsed = regex['meet_date'].findall(date_str)
    if len(date_parsed) == 0:
        raise ValueError(f'{date_str} cannot be parsed into the regex')
    elif len(date_parsed) == 1:
        date_parsed = date_parsed[0]
        yr = int(date_parsed[3])
        mon = date_parsed[0]
        date = int(date_parsed[1]) 
        return datetime.datetime.strptime(f'{yr:04}-{mon}-{date:02}', '%Y-%B-%d')

def insert_athlete(url):
    """"""

def parse_event(name, body, date):
    """Parse the event, create corresponding ORM objects and and update the database"""
    results = []
    if regex['5000m'].search(name):
        event = orm.Event(orm.EventType._5000m, date)
        session.add(event)
        session.flush()
        for row in body:
            pl = row.td[0]
            time = row.find_all('td')[4]
            athlete_id = regex['athlete_id'].match(row.a.href).group(1)
            if not session.get(orm.Athlete, athlete_id):
                insert_athlete(requests.get(row.a.href, headers=headers))
            results.append(orm.Result(event.id, athlete_id, time=time, place=pl))
            
    
async def scrape_meet(url):
    """Scrape an entire meet give the root meet page"""
    # scrape relevant information from url
    meet = requests.get(url, headers=headers)

    meet = bs(meet.content, features='xml')
    events = meet.find_all('div', {'class': 'row'})
    
    for ev in events[:1]:
        parse_event(ev.h3.text, ev.tbody)


    async with lock:
        # add all future scrapes to the queue
        pass 

def scrape_root(deadline):
    # root tfrrs call
    root = requests.get('https://www.tfrrs.org/results.rss', headers=headers)
    # schedule scraping events according to a uniform distribution over the time to scraping the next rss feed

    root = bs(root.content, features='xml')
    for meet in root.rss.channel.find_all('item', recursive=False):
        meet_title = meet.title.string
        meet_date = parse_date(meet.description.string)
        url = meet.link.string
        meet_id = regex['meet_results'].match(url).group(1)
        delay = random.uniform(DAY/24) # start all scraping tasks within an hour of the root scrape
        pending_scrapes.enter(delay, 1, scrape_meet, url)
        # insert into database TODO:uncomment when the structure is working!
        # cur.execute("INSERT INTO TABLE meets(id, name, date) VALUES(%s, %s)", (id, title, date))




pending_scrapes = sched.scheduler(time.time, time.sleep)
lock = asyncio.Lock()

# main loop
while True:
    now = time.time()
    pending_scrapes.enter(0, 1, scrape_root, now + DAY)
    time.sleep(DAY)
    


    


# 2. Go to the men's and women's events compiled pages
# 3. This is where I assume the logic gets a bit complicated: parse the event name at the top of each table, then loop through each of the entries and record names, There are heat tables under each of these, which I assume can be recorded to give more granular results
# If there are links associated with the athletes and schools, these should be used to populate the athlete id and school id fields
# when there is a school id that has not previously been seen, go to the school page and record its information. Wait, no maybe we should maintain a queue of pages to visit
# there are certain event types that I think will be challenging to parse. I think we should leave these for 
# we will need to write a text processor for the event names though. This shouldn't be too hard