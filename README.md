# Bactic
Latent Growth Models (and other statistical analyses) of the TFRRS database

## Overview
If you're an athlete or sports nerd, you have probably spent hours poring over statistics, looking at your performance history and obsessing over trivia otherwise boring to the casual onlooker. These meaningless numbers are actually a gold mine of insight and analysis that should excite the mathematically-inclined. 

TFRRS is a large online database for collegiate and other DirectAthletics-associated track and cross country events in the United States. It covers over 100 athletic conferences and reports tens of thousands of results on a weekly basis during season, making it a rich source of performance data.

In this project, we scrape TFRRS on a rolling basis, constructing a relational database of every event, heat, and recorded time on the day it is published. In a separate task, we run a number of statistical tests on this database. Some of these are just computing summary statistics and histograms, which are nonetheless very interesting to look at. Others are more sophisticated tests, such as the fitting of latent growth models to the time series of an athletes' performances. The final presentation of these statistics is given in a public website that comprises the third branch of this project's development.

## How is this project structured?
This project comprises a number of communicating services that run in a local docker orchestration or in an AWS deployment. Since the bulk of the project backend is written in Go, the directory structure tends to follow the standard Go project layout.

 - Scraper [Go]: service that runs a scraping job every 24 hours to populate the relational database with new event data
 - Stats API [Go]: hit this api to get stats
 - Stats engine [Python]: this engine computes stats at the API's demand and caches them in the stats cache
    - Note: this in the only component of the project that does not fall into the Go project structure. It resides in the `/stats`
 - Site [HTML/CSS]: although this is sacrilege nowadays, we are writing this in straight html and css
 - Relational Database [PostgreSQL]: stores all relational performance data from the scraper
 - Stats Cache [Redis]: caches computed statistics for quick access over a set time interval

## How can I use the data in this project?
The data scraped from DirectAthletics' TFRRS database falls under their [Terms of Use Policy](https://www.directathletics.com/terms_of_use.html), which states that any commercial reproduction of their data is prohibited. Basically, users are prohibited from selling or otherwise producing derivatives of this data for their own profit. Since this service is not a direct reproduction of TFRRS data and instead computes higher-order statistics and summaries that their service does not provide, it also does not pose as a competitor to their product. If there are any further questions about the legal nature of this project, please feel free to contact one of us.
