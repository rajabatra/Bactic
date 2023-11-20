package tfrrs

import (
	"bactic/internal"
	"bactic/internal/database"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/xmlquery"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func setupRSSCollector(rootCollector *colly.Collector, meetCollector *colly.Collector, db *database.BacticDB) {
	logger := log.New(os.Stdout, "XML RSS", log.LUTC)
	logger.SetPrefix("XML Root Collector")

	rootCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Root collector looking at meet RSS", r.URL)
	})
	colly.AllowURLRevisit()(rootCollector)

	rootCollector.OnXML("//item", func(x *colly.XMLElement) {
		node := x.DOM.(*xmlquery.Node)
		t := xmlquery.Find(node, "/title")
		d := xmlquery.Find(node, "/description")
		l := xmlquery.Find(node, "/link")
		if len(t) != 1 || len(d) != 1 || len(l) != 1 {
			logger.Println("Encountered malformed xml item for meet, not scraping")
			return
		}

		link := strings.TrimSpace(l[0].InnerText())
		date, err := parseMeetDate(strings.TrimSpace(d[0].InnerText()))
		title := strings.TrimSpace(t[0].InnerText())
		if err != nil {
			logger.Printf("Unable to parse date string, for meet %s, skipping", title)
			return
		}

		if err = db.InsertMeet(internal.Meet{
			ID:   uuid.New().ID(),
			Name: title,
			Date: date,
		}); err != nil {
			panic(err)
		}

		meetCollector.Visit(link)
	})
}

func setupSchoolCollector(schoolCollector *colly.Collector, db *database.BacticDB) {
	logger := log.Default()
	logger.SetPrefix("School Collector")

	schoolCollector.OnRequest(func(r *colly.Request) {
		logger.Println("School collector visiting school", r.URL)
	})

	schoolCollector.OnHTML("h3#team-name", func(h *colly.HTMLElement) {
		titleCaser := cases.Title(language.AmericanEnglish)
		teamName := titleCaser.String(strings.TrimSpace(h.Text))
		division := -1
		var leagues []string
		h.DOM.Parent().Siblings().First().Find("span.panel-heading-normal-text").First().Children().Each(func(i int, s *goquery.Selection) {
			d := parseDivision(s.Text())
			if division >= 0 && d >= 0 && division != d {
				logger.Fatalf("Found conflicting divisions in the parsed division list: %d, %d", division, d)
			} else if d >= 0 {
				division = d
			} else {
				leagues = append(leagues, strings.TrimSpace(s.Text()))
			}
		})

		if division < 0 {
			logger.Println("Could not parse a division from the school page", teamName)
		}

		school := internal.School{
			Name:     teamName,
			URL:      h.Request.URL.String(),
			Division: division,
			Leagues:  leagues,
		}
		_, err := db.InsertSchool(school)
		if err != nil {
			logger.Fatal("Error inserting school", err)
		}
	})
}

func setupAthleteCollector(athleteCollector *colly.Collector, db *database.BacticDB) {
	logger := log.Default()
	logger.SetPrefix("Athlete Collector")

	athleteCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Athlete collector visiting athlete", r.URL)
	})

	athleteCollector.OnHTML("h3.panel-title.large-title", func(e *colly.HTMLElement) {
		c := cases.Title(language.AmericanEnglish)
		athName := c.String(strings.Split(strings.TrimSpace(e.Text), "\n")[0])

		/*
			Athlete ID is a tough case because athlete name is not a unique identifier.
			We instead need to use the tfrrs id mapping to verify our own ids. TFRRS
			has link ids that map to a global id. This relation can be many-to-one,
			meaning we need to create an id of our own and verify that the link ids
			map to it before inserting the athlete into our database. By consistency,
			we define our own global ids for our athletes, meaning we have an additional
			mapping tfrrs id->bactic id which is one-to-one.

			Overall, the process of insertion is as follows:
			1. verify that the link id maps to a global id in the mapping.
			2. if so, get the global id from the mapping and return the mapped tfrrs id.
				You are done
			3. if not, you check to see if the global id is mapped in the table
			4. if not, this is a new athlete. Create a new tfrrs id, map the link
				to the global and then the global to the tfrrs id
			5. if so, then this is not a new athlete. Add the mapped to global relation
			in the mapping and then follow the global to tfrrs relation
		*/
		linkID, err := parseAthleteIDFromURL(e.Request.URL.String())
		if err != nil {
			panic(err)
		}
		TFRRSID, err := parseAthleteIDFromURL(e.Response.Request.URL.String())
		if err != nil {
			panic(err)
		}

		_, found := db.GetAthleteRelation(TFRRSID)
		if !found {
			global := uuid.New().ID()
			if err = db.AddAthleteRelation(linkID, global); err != nil {
				panic(err)
			}
			db.InsertAthlete(internal.Athlete{
				ID:   global,
				Name: athName,
			})
		} else {
			if err = db.AddAthleteRelation(linkID, TFRRSID); err != nil {
				panic(err)
			}
		}
	})
}

// func NewTFRRSXCCollector(db *database.BacticDB, meetID uint32) *colly.Collector {
// 	logger := log.New(os.Stdout, "XC Collector", log.LUTC)

// 	meetCollector := colly.NewCollector()
// 	schoolCollector := colly.NewCollector()
// 	athleteCollector := colly.NewCollector()
// 	setupAthleteCollector(athleteCollector, db)
// 	setupSchoolCollector(schoolCollector, db)

// 	meetCollector.OnRequest(func(r *colly.Request) {
// 		logger.Println("Visiting meet", r.URL)
// 	})

// 	meetCollector.OnHTML("div.row", func(h *colly.HTMLElement) {
// 		resultsRows := h.DOM.Find("tbody>tr")
// 		tableLength := resultsRows.Length()
// 		if tableLength == 0 {
// 			return
// 		}
// 		header := strings.ToLower(h.DOM.Find("div.custom-table-title-xc>h3").Text())
// 		// we do not parse team results
// 		if strings.Contains(header, "team results") {
// 			return
// 		}

// 		eventType, err := parseXCEventType(h.DOM.Find("div.custom-table-title-xc>h3").Text())
// 		if err != nil {
// 			logger.Fatal("Unable to parse this table type, which should not happen in XC", err)
// 		}

// 		rowLength := resultsRows.First().Children().Length()
// 		table := make([][][]string, tableLength)

// 		resultsRows.Each(func(i int, s *goquery.Selection) {
// 			table[i] = make([][]string, rowLength)
// 			s.Children().Each(func(j int, r *goquery.Selection) {
// 				// strip text and link if it exists
// 				table[i][j] = make([]string, 0, 2)
// 				table[i][j] = append(table[i][j], strings.TrimSpace(r.Text()))
// 				href, found := r.Children().Attr("href")
// 				if found {
// 					table[i][j] = append(table[i][j], href)
// 				}
// 			})
// 		})
// 		// parse all information from table
// 		resultTable, athleteURLs, schoolURLs := parseResultTable(table, logger, eventType)
// 		logger.Printf("%v, %v", resultTable, eventType)
// 		athletesToScrape := db.GetMissingAthletes(resultTable)
// 		schoolsToScrape := db.GetMissingSchools(schoolURLs)

// 		// visit schools before athletes due to the database relation dependencies
// 		for _, url := range schoolsToScrape {
// 			schoolCollector.Visit(url)
// 		}

// 		for _, athleteID := range athletesToScrape {
// 			athleteCollector.Visit(athleteURLs[athleteID])
// 		}

// 		// populate the athlete-school relations
// 		for i, url := range schoolURLs {
// 			school, found := db.GetSchoolURL(url)
// 			if !found {
// 				logger.Fatal("We must be able to find the school", url)
// 			}
// 			db.AddAthleteToSchool(resultTable[i].AthleteID, school.ID)
// 		}

// 		// once this is done, we can insert the heat
// 		_, err = db.InsertHeat(eventType, meetID, resultTable)
// 		if err != nil {
// 			panic(err)
// 		}
// 	})
// 	return meetCollector
// }

func NewTFRRSTrackCollector(db *database.BacticDB, meetID uint32) *colly.Collector {

	logger := log.Default()
	logger.SetPrefix("Meet Collector")

	meetCollector := colly.NewCollector()
	schoolCollector := colly.NewCollector()
	athleteCollector := colly.NewCollector()
	setupAthleteCollector(athleteCollector, db)
	setupSchoolCollector(schoolCollector, db)

	meetCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Visiting meet", r.URL)
	})

	meetCollector.OnHTML("div.row", func(e *colly.HTMLElement) {
		resultsRows := e.DOM.Find("tbody>tr")
		tableLength := resultsRows.Length()
		if tableLength == 0 {
			return
		}

		// the header should be present under other circumstances
		eventType, err := parseEvent(e.DOM.Find("div.custom-table-title>h3").Text())
		if err != nil {
			logger.Println("Unable to parse this table type. Assuming a redundant heat table:", err)
			return
		}

		rowLength := resultsRows.First().Children().Length()
		table := make([][][]string, tableLength)

		// Collect into table struct
		resultsRows.Each(func(i int, s *goquery.Selection) {
			table[i] = make([][]string, rowLength)
			s.Children().Each(func(j int, r *goquery.Selection) {
				// strip text and link if it exists
				table[i][j] = make([]string, 0, 2)
				table[i][j] = append(table[i][j], strings.TrimSpace(r.Text()))
				href, found := r.Children().Attr("href")
				if found {
					table[i][j] = append(table[i][j], href)
				}
			})
		})

		// parse all information from table
		resultTable, athleteIDs, schoolURLs := parseResultTable(table, logger, eventType)
		globalIDs := db.GetMissingAthletes(athleteIDs)
		// follow the url and populate if the globalID is nil
		for i, id := range globalIDs {
			if id == nil {
				athleteCollector.Visit(fmt.Sprintf("https://www.tfrrs.org/athletes/%v", athleteIDs[i]))
				global, found := db.GetAthleteRelation(athleteIDs[i])
				if !found {
					panic("The key should have a global entry in the table")
				}
				resultTable[i].AthleteID = global
			} else {
				resultTable[i].AthleteID = *id
			}
		}
		schoolsToScrape := db.GetMissingSchools(schoolURLs)

		// visit schools before athletes due to the database relation dependencies
		for _, url := range schoolsToScrape {
			schoolCollector.Visit(url)
		}

		// populate the athlete-school relations
		for i, url := range schoolURLs {
			school, found := db.GetSchoolURL(url)
			if !found {
				logger.Fatal("We must be able to find the school", url)
			}
			db.AddAthleteToSchool(resultTable[i].AthleteID, school.ID)
		}

		// once this is done, we can insert the heat
		_, err = db.InsertHeat(eventType, meetID, resultTable)
		if err != nil {
			panic(err)
		}
	})
	return meetCollector
}
