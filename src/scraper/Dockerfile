FROM python:3.11

WORKDIR /usr/src/app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY ./tfrrs_scraper.py .

CMD ["python3" "./tfrrs_scraper.py"]