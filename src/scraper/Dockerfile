FROM python:3.11

WORKDIR /usr/src/app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY ./tfrrs_scraper.py .
# TODO: set this up as a cron task. Try to add some form of variation so we don't run at the same exact time  on a daily schedule
CMD [ "python", "./tfrrs_scraper.py" ]