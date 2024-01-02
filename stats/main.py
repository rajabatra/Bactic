import psycopg
import redis
import os
import numpy as np
import http
from typing import List
import datetime

db_uri = os.getenv('DB_URL')
redis_uri = os.getenv('CACHE_URI')
if not db_uri:
    raise ValueError("Could not find environment variable DB_URL")

db = psycopg.connect(db_uri)
# cache = redis.Redis(redis_uri)

cur = db.cursor()

# just run a simple query on the database to assure we have a connection
def hist(buckets: int, event_type: int, cur):
    cur.execute("SELECT quant FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.event_type = %s", (event_type,))
    results = list(map(lambda x: x[0], cur.fetchall()))
    fig, ax = plt.subplots()
    def timeTicks(x, pos):
        d = datetime.timedelta(seconds=x)                                                                                                                                                                                                                                          
        return str(d) 
    formatter = matplotlib.ticker.FuncFormatter(timeTicks)
    ax.hist(results, 150)
    ax.xaxis.set_major_formatter(formatter)
    ax.set_xlabel('Time')
    ax.set_ylabel('Count')
    ax.set_title("Men's 10k times since November 16")
    plt.show()
    # return np.histogram(results, bins=buckets)

# compute this statistic for cross country 8k
hist(100, 26, cur)
# print(hist(100, 26, cur))


# Heres the plan:
# First get a framework of computing stats together. We are going to have a lot of stats
# it makes sense to learn how to modularize them

db.close()
