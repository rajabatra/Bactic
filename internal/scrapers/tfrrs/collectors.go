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

// Create a new collector that reads
func NewRSSCollector(db *database.BacticDB) *colly.Collector {
	rootCollector := colly.NewCollector(colly.AllowURLRevisit())
	//TODO: we may want to just create one collector eventually
	xcCollector := NewTFRRSXCCollector(db)
	tfCollector := NewTFRRSTrackCollector(db)

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
		meetID := uuid.New().ID()

		if err = db.InsertMeet(internal.Meet{
			ID:   meetID,
			Name: title,
			Date: date,
		}); err != nil {
			panic(err)
		}
		ctx := colly.NewContext()
		ctx.Put("MeetID", meetID)
		if strings.Contains(link, "results/xc/") {
			if err := xcCollector.Request("GET", link, nil, ctx, nil); err != nil {
				panic(err)
			}
		} else if strings.Contains(link, "results/tf/") {
			if err := tfCollector.Request("GET", link, nil, ctx, nil); err != nil {
				panic(err)
			}
		} else {
			panic("Unable to classify meet link as either tf or results")
		}

		rootCollector.Visit(link)
	})

	return rootCollector
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
		linkID, found := e.Request.Ctx.GetAny("linkID").(uint32)
		if !found {
			panic("The value linkID not found in request context")
		}

		tfrrsID, err := parseAthleteIDFromURL(e.Request.URL.String())
		if err != nil {
			panic(err)
		}

		_, found = db.GetAthleteRelation(tfrrsID)
		if found {
			if err := db.AddAthleteRelation(linkID, tfrrsID); err != nil {
				panic(err)
			}
			return
		}

		bacticID := uuid.New().ID()
		if err := db.AddAthleteRelation(tfrrsID, bacticID); err != nil {
			panic(err)
		}
		if linkID != tfrrsID {
			if err := db.AddAthleteRelation(linkID, tfrrsID); err != nil {
				panic(err)
			}
		}

		c := cases.Title(language.AmericanEnglish)
		athName := c.String(strings.Split(strings.TrimSpace(e.Text), "\n")[0])

		if err := db.InsertAthlete(internal.Athlete{
			ID:   bacticID,
			Name: athName,
		}); err != nil {
			panic(err)
		}
	})
}

func NewTFRRSXCCollector(db *database.BacticDB) *colly.Collector {
	logger := log.New(os.Stdout, "XC Collector", log.LUTC)

	meetCollector := colly.NewCollector()
	schoolCollector := colly.NewCollector()
	athleteCollector := colly.NewCollector()
	setupAthleteCollector(athleteCollector, db)
	setupSchoolCollector(schoolCollector, db)

	meetCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Visiting meet", r.URL)
	})

	meetCollector.OnHTML("div.row", func(h *colly.HTMLElement) {
		resultsRows := h.DOM.Find("tbody>tr")
		tableLength := resultsRows.Length()
		if tableLength == 0 {
			return
		}
		header := strings.ToLower(strings.Split(h.DOM.Find("div.custom-table-title-xc>h3").Text(), "\n")[0])
		// TODO: we do not parse team results for now
		if strings.Contains(header, "team results") {
			return
		}

		eventType, err := parseXCEventType(header)
		if err != nil {
			return
		}

		rowLength := resultsRows.First().Children().Length()
		table := make([][][]string, tableLength)

		// Collect table into struct indexed by row, column, (text, href)
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
		resultTable, linkIDs, schoolURLs := parseResultTable(table, logger, eventType)

		for _, link := range linkIDs {
			_, found := db.GetTFRRSAthleteID(link)
			if !found {
				ctx := colly.NewContext()
				ctx.Put("linkID", link)
				athleteCollector.Request("GET", fmt.Sprintf("https://www.tfrrs.org/athletes/%v", link), nil, ctx, nil)
			}
		}

		// second pass, back-populate
		for i, link := range linkIDs {
			tfrrsID, found := db.GetTFRRSAthleteID(link)
			if found {
				resultTable[i].AthleteID = tfrrsID
			} else {
				panic("Should be able to find tfrrs id after first pass")
			}
		}
		schoolsToScrape := db.GetMissingSchools(schoolURLs)

		for _, url := range schoolsToScrape {
			schoolCollector.Visit(url)
		}

		// populate athlete-school relations
		for i, url := range schoolURLs {
			school, found := db.GetSchoolURL(url)
			if !found {
				logger.Fatal("We must be able to find the school", url)
			}
			if err := db.AddAthleteToSchool(resultTable[i].AthleteID, school.ID); err != nil {
				panic(err)
			}
		}

		// finally, insert the heat
		meetID := h.Request.Ctx.GetAny("MeetID").(uint32)
		_, err = db.InsertHeat(eventType, meetID, resultTable)
		if err != nil {
			panic(err)
		}
	})
	return meetCollector
}

func NewTFRRSTrackCollector(db *database.BacticDB) *colly.Collector {
	logger := log.New(os.Stdout, "TF Collector", log.LUTC)

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

		/*
			Athlete ID is a tough case because an athletes name is not a unique identifier.
			We instead need to use the tfrrs id mapping to verify our own ids. TFRRS
			has link ids that map to a global id associated with the page of each athlete.
			The permanence of this id is currently unknown, however I suspect that this is not the case.
			This will be an additional issue we need to tackle if it comes up.
			This relation can be many-to-one,
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
		resultTable, linkIDs, schoolURLs := parseResultTable(table, logger, eventType)

		// first pass
		for _, link := range linkIDs {
			_, found := db.GetTFRRSAthleteID(link)
			if !found {
				ctx := colly.NewContext()
				ctx.Put("linkID", link)
				athleteCollector.Request("GET", fmt.Sprintf("https://www.tfrrs.org/athletes/%v", link), nil, ctx, nil)
			}
		}

		// second pass, back-populate
		for i, link := range linkIDs {
			tfrrsID, found := db.GetTFRRSAthleteID(link)
			if found {
				resultTable[i].AthleteID = tfrrsID
			} else {
				panic("Should be able to find tfrrs id after first pass")
			}
		}
		schoolsToScrape := db.GetMissingSchools(schoolURLs)

		// visit schools before athletes due to the database relational dependencies
		for _, url := range schoolsToScrape {
			schoolCollector.Visit(url)
		}

		// populate the athlete-school relations
		for i, url := range schoolURLs {
			school, found := db.GetSchoolURL(url)
			if !found {
				logger.Fatal("We must be able to find the school", url)
			}
			if err := db.AddAthleteToSchool(resultTable[i].AthleteID, school.ID); err != nil {
				panic(err)
			}
		}

		// once this is done, we can insert the heat
		meetID := e.Request.Ctx.GetAny("MeetID").(uint32)
		_, err = db.InsertHeat(eventType, meetID, resultTable)
		if err != nil {
			panic(err)
		}
	})
	return meetCollector
}
