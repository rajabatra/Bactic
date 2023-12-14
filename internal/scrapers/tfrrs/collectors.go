package tfrrs

import (
	"bactic/internal"
	"bactic/internal/database"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
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
func NewRSSCollector(db *sql.DB, ctx context.Context) *colly.Collector {
	rootCollector := colly.NewCollector(colly.AllowURLRevisit())
	meetCollector := NewMeetCollector(ctx)

	logger := log.New(os.Stdout, "XML RSS", log.LUTC)
	logger.SetPrefix("XML Root Collector")

	rootCollector.OnRequest(func(r *colly.Request) {
		logger.Println("Root collector looking at meet RSS", r.URL)
	})
	colly.AllowURLRevisit()(rootCollector)

	rootCollector.OnXML("//item", func(x *colly.XMLElement) {
		select {
		// ignore all xml meets when cancelled
		case <-ctx.Done():
			return
		default:

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

			tx, err := db.Begin()
			if err != nil {
				panic(err)
			}

			if err = database.InsertMeet(tx, internal.Meet{
				ID:   meetID,
				Name: title,
				Date: date,
			}); err != nil {
				panic(err)
			}

			meetCtx := colly.NewContext()
			meetCtx.Put("MeetID", meetID)
			meetCtx.Put("tx", tx)
			if err := meetCollector.Request("GET", link, nil, meetCtx, nil); err != nil {
				panic(err)
			}

			// if we have cancelled, do not insert
			select {
			case <-ctx.Done():
				return
			default:
				if err := tx.Commit(); err != nil {
					panic(err)
				}
			}
		}
	})
	return rootCollector
}

func NewMeetCollector(ctx context.Context) *colly.Collector {
	logger := log.New(os.Stdout, "Meet Collector ", log.Ldate|log.Ltime)

	meetCollector := colly.NewCollector()

	meetCollector.OnRequest(func(r *colly.Request) {
		logger.Println("visiting meet", r.URL)
	})

	meetCollector.OnHTML("body > div.page.container > div > div > div.panel-second-title > div > div.col-lg-8 > div:nth-child(2) > span.panel-heading-normal-text", func(h *colly.HTMLElement) {
		h.DOM.Children().Each(func(i int, s *goquery.Selection) {
			link, exists := s.Attr("href")
			if !exists {
				panic("A link to the track meet should exist")
			}
			h.Request.Visit(link)
		})
	})

	meetCollector.OnHTML("div.row", func(h *colly.HTMLElement) {
		tx := h.Request.Ctx.GetAny("tx").(*sql.Tx)
		resultsRows := h.DOM.Find("tbody>tr")
		tableLength := resultsRows.Length()
		if tableLength == 0 {
			return
		}

		url := h.Request.URL.String()
		var (
			eventType internal.EventType
			err       error
		)

		if strings.Contains(url, "/xc/") {
			header := strings.ToLower(strings.Split(h.DOM.Find("div.custom-table-title-xc>h3").Text(), "\n")[0])
			// TODO: we do not parse team results for now
			if strings.Contains(header, "team results") {
				return
			}

			eventType, err = parseXCEventType(header)
			if err != nil {
				return
			}
		} else { // assume tf otherwise
			eventType, err = parseEvent(h.DOM.Find("div.custom-table-title>h3").Text())
			if err != nil {
				logger.Println("Unable to parse this table type. Assuming a redundant heat table:", err)
				return
			}
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
		// parse all information from table
		resultTable, linkIDs, schoolURLs := parseResultTable(table, logger, eventType)
		validResults := make([]internal.Result, 0)

		for i, link := range linkIDs {
			select {
			// this is the expensive step of the page scrape, meaning we cancel here when context.Done channel is closed
			case <-ctx.Done():
				return
			default:
				id, err, httpError := checkAthlete(tx, link, logger)
				if err != nil {
					panic(err)
				} else if httpError {
					continue
				}
				resultTable[i].AthleteID = id
				validResults = append(validResults, resultTable[i])

				school := checkSchool(tx, schoolURLs[i], logger)
				if err := database.AddAthleteToSchool(tx, resultTable[i].AthleteID, school.ID); err != nil {
					logger.Panic(err, school.ID)
				}
			}
		}

		// finally, insert the heat
		meetID := h.Request.Ctx.GetAny("MeetID").(uint32)
		_, err = database.InsertHeat(tx, eventType, meetID, validResults)
		if err != nil {
			panic(err)
		}
	})
	return meetCollector
}

// How we scrape athletes since there are some scraping dependencies that are challenging to handle through collys functional scraping mechanisms
func checkAthlete(tx *sql.Tx, linkID uint32, logger *log.Logger) (athleteID uint32, err error, httpError bool) {
	tfrrsID, found := database.GetAthleteRelation(tx, linkID)
	// if the link ID is in the table, we return what we find directly
	if found {
		bacticID, found := database.GetAthleteRelation(tx, tfrrsID)
		if !found {
			return tfrrsID, nil, false
		} else {
			return bacticID, nil, false
		}
	}

	// otherwise, we follow the link to validate the tfrrs id
	resp, err := http.Get(fmt.Sprintf("https://www.tfrrs.org/athletes/%v", linkID))
	if err != nil {
		panic(err)
	} else if resp.StatusCode > 400 {
		return 0, nil, true
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	tfrrsID, err = parseAthleteIDFromURL(resp.Request.URL.String())
	if err != nil {
		panic(err)
	}

	// we have a new reference to the same tfrrs id
	bacticID, found := database.GetAthleteRelation(tx, tfrrsID)
	if found {
		if err = database.AddAthleteRelation(tx, linkID, tfrrsID); err != nil {
			panic(err)
		}
		return bacticID, nil, false
	}

	// we have to create a new athlete
	bacticID = uuid.New().ID()
	if err := database.AddAthleteRelation(tx, tfrrsID, bacticID); err != nil {
		panic(err)
	}
	if linkID != tfrrsID {
		if err := database.AddAthleteRelation(tx, linkID, tfrrsID); err != nil {
			panic(err)
		}
	}

	c := cases.Title(language.AmericanEnglish)
	h := doc.Selection.Find("h3.panel-title.large-title")
	athFields := strings.Fields(h.Text())
	athFields = athFields[:len(athFields)-1]
	athName := c.String(strings.Join(athFields, " "))
	logger.Printf("Found new athlete %s, scraping", athName)

	if err := database.InsertAthlete(tx, internal.Athlete{
		ID:   bacticID,
		Name: athName,
	}); err != nil {
		panic(err)
	}
	return bacticID, nil, false
}

// checks the url string for existence. If not, scrape the school and then insert. Otherwise, insert the school
func checkSchool(tx *sql.Tx, url string, logger *log.Logger) internal.School {
	school, found := database.GetSchoolURL(tx, url)
	if found {
		return school
	}

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	nameDiv := doc.Selection.Find("h3#team-name")
	titleCaser := cases.Title(language.AmericanEnglish)
	teamName := titleCaser.String(strings.TrimSpace(nameDiv.Text()))

	division := -1
	var leagues []string

	nameDiv.Parent().Siblings().First().Find("span.panel-heading-normal-text").First().Children().Each(func(i int, s *goquery.Selection) {
		d := parseDivision(s.Text())
		if division >= 0 && d >= 0 && division != d {
			log.Fatalf("Found conflicting divisions in the parsed division list: %d, %d", division, d)
		} else if d >= 0 {
			division = d
		} else {
			leagues = append(leagues, strings.TrimSpace(s.Text()))
		}
	})

	if division < 0 {
		logger.Println("Could not parse a division from the school page", teamName)
	}

	school = internal.School{
		ID:       uuid.New().ID(),
		Name:     teamName,
		URL:      url,
		Division: division,
		Leagues:  leagues,
	}

	err = database.InsertSchool(tx, school)
	if err != nil {
		panic(err)
	}
	return school
}
