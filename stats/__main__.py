import psycopg2
import redis
import os
import numpy as np
from typing import List

db_url = os.getenv('DB_URL')
if not db_url:
    raise ValueError("Could not find environment variable DB_URL")

conn = psycopg2.connect(db_url)

cur = conn.cursor()

# just run a simple query on the database to assure we have a connection
def hist(buckets: int, event_type: int, cur) -> np.array:
    h = []
    cur.execute("SELECT quant FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.event_type = %s", (event_type,))
    results = cur.fetchall()
    return np.histogram(results, bins=buckets)

# compute this statistic for cross country 8k
print(hist(3, 26, cur))

conn.close()
