import requests, csv, os
from bs4 import BeautifulSoup


# initialize db connection
dbname = os.getenv("POSTGRES_DB")
dbuser = os.getenv("POSTGRES_USER")
dbpass = os.getenv("POSTGRES_PASSWORD")

conn = psycopg2.connect(dbname=dbname, user=dbuser, password=dbpass)


# # URL of the website to be scraped
# url = "https://www.athletic.net/TrackAndField/AthleteRecords.aspx?SchoolID=1928"

# # Send a request to the website and get its content
# response = requests.get(url)
# content = response.content

# # Parse the HTML content using BeautifulSoup
# soup = BeautifulSoup(content, "html.parser")

# # Find the table element that contains the names
# # obtain information from html tag <table>
# div_ele = soup.find('div', {"class": "tab-content row"})
# table = div_ele.find('table')
# table


# # Extract the names, distances, and times from the table
# names = []
# distances = []
# times = []
# for row in table.find_all("tr"):
#     # Skip the header row
#     if "th" in row.find_all():
#         continue
#     # Extract the name, distance, and time from the row
#     cells = row.find_all("td")
#     name = cells[0].get_text().strip()
#     distance = cells[0].get_text().strip()
#     time = cells[0].get_text().strip()
#     names.append(name)
#     distances.append(distance)
#     times.append(time)

# # Create a CSV file and write the names, distances, and times to it
# with open("athletes.csv", mode="w", newline="") as file:
#     writer = csv.writer(file)
#     writer.writerow(["Name", "Distance", "Time"])
#     for i in range(len(names)):
#         writer.writerow([names[i], distances[i], times[i]])