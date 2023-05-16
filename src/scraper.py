from bs4 import BeautifulSoup as bs
import re, queue, datetime, os, requests, psycopg2, random

postgres_pass = os.environ['POSTGRES_PASSWORD']

n_requests = int(1e3)

# initialize database connection
db = psycopg2.connect(f'dbname=bactic user=postgres password={postgres_pass}')
cur = db.cursor()

# we have some logic here that can be the general framework for our scraping strategy

# create some sort of request randomization scheme
request_intervals_sec = [100*random.random() for _ in range(n_requests)]

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

# root tfrrs call
root = requests.get('https://www.tfrrs.org/results.rss', headers=headers)

# all the regexes we want to use
regex = {
    'meet_results': re.compile('https://www.tfrrs.org/results/(\d+)'),
    'meet_date': re.compile('([a-zA-Z]+)\s*(\d+)(\s*-\s*\d+)?,\s*(\d{4})')
}

# a queue to keep track of data that we intend to eventually scrape
pending_scrapes = queue.PriorityQueue()

root = bs(root.content, features='xml')
for meet in root.rss.channel.find_all('item', recursive=False):
    title = meet.title.string
    date = parse_date(meet.description.string)
    url = meet.link.string
    id = regex['meet_results'].findall(url)[0]
    # insert into database
    cur.execute("INSERT INTO TABLE meets(id, name, date) VALUES(%s, %s)", (id, title, date))

def scrape_meet(url):
    pass


    


# 2. Go to the men's and women's events compiled pages
# 3. This is where I assume the logic gets a bit complicated: parse the event name at the top of each table, then loop through each of the entries and record names, There are heat tables under each of these, which I assume can be recorded to give more granular results
# If there are links associated with the athletes and schools, these should be used to populate the athlete id and school id fields
# when there is a school id that has not previously been seen, go to the school page and record its information. Wait, no maybe we should maintain a queue of pages to visit
# there are certain event types that I think will be challenging to parse. I think we should leave these for 
# we will need to write a text processor for the event names though. This shouldn't be too hard