from sqlalchemy import create_engine
from sqlalchemy.orm import Session
import datetime
import orm
import unittest
import os, requests, bs4

import tfrrs_scraper


# root_rss = tfrrs_scraper.get_bs('https://www.tfrrs.org/results.rss')
# example_meet = tfrrs_scraper.get_bs('https://www.tfrrs.org/results/79700/m/2023_SCIAC_TF_Championships')
# example_athlete = tfrrs_scraper.get_bs('https://www.tfrrs.org/athletes/7374205/Claremont_Mudd_Scripps/Henry_Pick.html')

tmp_dir = './tmp'

def load_and_save(url, tmp_directory):
    fname = url.split('/')[-1]
    if not os.path.exists(tmp_directory):
        os.mkdir(tmp_directory)
    
    if not os.path.exists(os.path.join(tmp_directory, fname)):
        html = requests.get(url).content
        file = open(os.path.join(tmp_directory, fname), 'wb')
        file.write(html)
    else:
        file = open(os.path.join(tmp_directory, fname), 'rb')
        html = file.read()

    file.close()
    
    return bs4.BeautifulSoup(html, features='lxml')
    

class TestDateParsing(unittest.TestCase):
    def test_parse_date(self):
        date = tfrrs_scraper.parse_date('May 20, 2023')
        expected = datetime.datetime(2023, 5, 20).date()
        self.assertEqual(date, expected)
    
    def test_parse_date_range(self):
        date = tfrrs_scraper.parse_date('May 12-14, 2023')
        expected = datetime.datetime(2023, 5, 12).date()
        self.assertEqual(date, expected)

class TestDivisionParsing(unittest.TestCase):
    def test_parse_division_diii(self):
        division = tfrrs_scraper.parse_division('DIII West Region')
        self.assertEqual(division, orm.Division.NCAADIII)

    def test_parse_division_dii(self):
        division = tfrrs_scraper.parse_division('DI Mountain Region')
        self.assertEqual(division, orm.Division.NCAADI)

    def test_parse_division_di(self):
        division = tfrrs_scraper.parse_division('DII South Central Region')
        self.assertEqual(division, orm.Division.NCAADII)

    def test_parse_division_naia(self):
        division = tfrrs_scraper.parse_division('NAIA')
        self.assertEqual(division, orm.Division.NAIA)

class TestTimeParsing(unittest.TestCase):
    def test_parse_time(self):
        time = tfrrs_scraper.parse_time('14:30.69')
        self.assertEqual(time, datetime.timedelta(0, 60*14 + 30.69))
    
    def test_parse_time_sub_minute(self):
        time = tfrrs_scraper.parse_time('10.98')
        self.assertEqual(time, datetime.timedelta(0, 10.98))
    
    def test_parse_time_error_handling(self):
        self.assertRaises(ValueError, tfrrs_scraper.parse_time, '1:02:15.98')

class TestEventBucketing(unittest.TestCase):
    def bucket_5k(self):
        event = tfrrs_scraper.bucket_event("Men's 5000 Meters")
        self.assertEqual(event, tfrrs_scraper.Event._5000m)
    
    def bucket_error(self):
        self.assertRaises(ValueError, tfrrs_scraper.bucket_event, "Women's competitive rowing")

class EventTableParsing(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        meet = load_and_save('https://www.tfrrs.org/results/79700/m/2023_SCIAC_TF_Championships.html', tmp_dir)
        events = meet.find_all('div', {'class': 'row'})
        cls._5k_event_table = events[24]

        cls.engine = create_engine("sqlite://", echo=False)
        orm.Base.metadata.create_all(cls.engine)

        cls.session = Session(cls.engine)
        return super().setUpClass()
    
    def test_parse_event_5k_table(self):
        tfrrs_scraper.parse_event(EventTableParsing._5k_event_table, orm.Sex.MALE, datetime.datetime(2023, 5, 5).date(), EventTableParsing.session)

    @classmethod
    def tearDownClass(cls) -> None:
        cls.session.close()
        orm.Base.metadata.drop_all(cls.engine)
        return super().tearDownClass()
    

class TestCheckAthlete(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.ath = load_and_save('https://www.tfrrs.org/athletes/7917933/Pomona_Pitzer/Lucas_Florsheim.html', tmp_dir)
        cls.engine = create_engine("sqlite://", echo=False)
        orm.Base.metadata.create_all(cls.engine)

        cls.session = Session(cls.engine)
        return super().setUpClass()
    
    def test_check_school(self):
        tfrrs_scraper.check_school('https://www.tfrrs.org/athletes/7917933/Pomona_Pitzer/Lucas_Florsheim.html')

    def test_check_athlete(self):
        tfrrs_scraper.check_school('')
    @classmethod
    def tearDownClass(cls) -> None:
        cls.session.close()
        orm.Base.metadata.drop_all(cls.engine)
        return super().tearDownClass()