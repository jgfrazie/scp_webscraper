package scraper

import (
	// "fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly"
	"github.com/jgfrazie/scp_webscraper/src/scpProcessor"
	"github.com/jgfrazie/scp_webscraper/src/utils"
)

// TODO: ADD A LOGGER FOR ALL ACCESS REQUESTS AND ERRORS AS TO NOT HAVE THE ERRORS APPEAR FOR THE USER

const (
	NumReg = "^Item #: [A-z]*[0-9]*"
	ObjReg = "^Object Class: [A-z]*"
	SpecialContainmentProceeduresReg = "^Special Containment Procedures: [A-z]*"
	DescriptionReg = "^Description: [A-z]*"

	BaseURL = "https://scp-wiki.wikidot.com"
	SeriesRegexp = "^/scp-series-*[0-9]*"
	SCPRegexp = "^/scp-[0-9]+"
)

// Use this when threading for multiple URL requests as to not get flagged by the server as an attack
const ServerResponseGuard = time.Millisecond * 50
var wg sync.WaitGroup


// Given a siteURL and a REGEXP as strings, will return all links on the page which satisify these requirements
// siteURL (string): URL of the site to scrape
// constraint (string): A REGEXP of specifications of the links to return
// RETURNS: *[]string: All available hrefs to be scraped
func LinkCollector(siteURL, constraint string) *[]string {
	c := colly.NewCollector()

	var links *[]string = &[]string{}


	// Passed to c.OnHTML to gather the links necessary to scrape
	collectorClosure := func(linkPointer *[]string) func(*colly.HTMLElement) {
		return func(e *colly.HTMLElement) {
			link := e.Attr("href")
			if match, _ := regexp.MatchString(constraint, link); match {
				*linkPointer = append(*linkPointer, link)
			}
		}
	}

	c.OnHTML("a", collectorClosure(links))

	// Error log put here in case in the future the website changes format
	c.OnError(func(_ *colly.Response, err error) {
		loggedErr := "Something went wrong: " + err.Error()
		log.Println(loggedErr)
		panic(loggedErr)
	})

	c.Visit(siteURL)

	return links
}

// Returns a 2D slice of all available SCP URLs to access
// ARGS:
//	availableSeries: All available series URLs for scraping
// RETURNS:
//	2D array of accessable SCP links
func CollectSCPs (availableSeries *[]string) *[]*[]string {
	return utils.Map(func(series string) *[]string {
	   return utils.Map(func(s string) string {
		   return BaseURL + s
		   }, LinkCollector(series, SCPRegexp))
	   }, availableSeries)
}

// Given an SCP-URL, will scrape all possible information from the webpage
// ARGS:
//	scpURL: A URL to the SCP page
// RETURNS:
//	An SCP struct
func SCPInfoCollector(scpURL string) *scpProcessor.SCPEntity {
	c := colly.NewCollector()
	var scp *scpProcessor.SCPEntity = &scpProcessor.SCPEntity{}
	scp.URL = scpURL

	// Passed to c.OnHTML to gather the links necessary to scrape
	collectorClosure := func() func(*colly.HTMLElement) {
		return func(e *colly.HTMLElement) {
			// Gets SCP number
			if ok, _ := regexp.MatchString(NumReg, e.Text); ok {
				if parsedStr := (strings.Split(e.Text, "SCP-")); len(parsedStr) != 2 {
					log.Println("unable to parse ItemNumber")
					scp.ItemNumber = 0
				} else {
					convertedNum, err := strconv.Atoi(parsedStr[1])
					if err != nil {
						log.Println(err)
						convertedNum = 0
					}

					scp.ItemNumber = convertedNum
				}
			}

			// Gets SCP Object Class
			if ok, _ := regexp.MatchString(ObjReg, e.Text); ok {
				var class string
				if parsedStr := (strings.Split(e.Text, ": ")); len(parsedStr) != 2 {
					class = "Undefined"	
				} else {
					class = string(parsedStr[1])
				}

				switch class {
				case "Safe":
					scp.ObjectClass = scpProcessor.Safe
				case "Euclid":
					scp.ObjectClass = scpProcessor.Euclid
				case "Keter":
					scp.ObjectClass = scpProcessor.Keter
				case "Thaumiel":
					scp.ObjectClass = scpProcessor.Thaumiel
				case "Neutralized":
					scp.ObjectClass = scpProcessor.Neutralized
				case "Decommissioned":
					scp.ObjectClass = scpProcessor.Decommissioned
				case "Apollyon":
					scp.ObjectClass = scpProcessor.Apollyon
				case "Archon":
					scp.ObjectClass = scpProcessor.Archon
				default:
					log.Println("unable to parse object class for SCP-", scp.ItemNumber)
					scp.ObjectClass = scpProcessor.Undefined
				}
			}

			// Get SCP containment Proceedures
			if ok, _ := regexp.MatchString(SpecialContainmentProceeduresReg, e.Text); ok {
				if parsedStr := strings.Split(e.Text, ": "); len(parsedStr) != 2 {
					scp.SpecialContainmentProceedures = "N/A"
				} else {
					scp.SpecialContainmentProceedures = parsedStr[1]
				}
			}

			// Get SCP description
			if ok, _ := regexp.MatchString(DescriptionReg, e.Text); ok {
				if parsedStr := strings.Split(e.Text, ": "); len(parsedStr) != 2 {
					scp.Description = "N/A"
				} else {
					scp.Description = parsedStr[1]
				}
			}
		}
	}

	c.OnHTML("p", collectorClosure())

	// Error log put here in case an entity page is not scrapable
	c.OnError(func(_ *colly.Response, err error) {
		loggedErr := "Something went wrong: " + err.Error()
		log.Println(loggedErr)
		panic(loggedErr)
	})

	c.Visit(scpURL)

	return scp
}

// Checks the error rate on a specific series of SCPs and returns a percent error rate
// An error is when any part of the SCPEntity struct is left as undefined
func ScraperErrorRate(allSCPs *[]*[]string, checkSeries int) float32 {
	var numErrors uint64	// The atomic counter for errors
	var numEntities uint64 = uint64(len(*allSCPs) * len(*((*allSCPs)[checkSeries])))

	// Checks if an SCP is fully scrapable. If not, then it is counted as an error
	checkSCP := func(scpURL string, errorCounter *uint64) {
		log.Printf("Gopher checking SCP-%v", scpURL[len(scpURL) - 4:])
		if scp := SCPInfoCollector(scpURL);
		scp.ItemNumber == 0 ||							// Could not get item number
		scp.ObjectClass == scpProcessor.Undefined ||	// Could not get object class
		scp.SpecialContainmentProceedures == "N/A" ||	// Could not get spec. con. pro.
		scp.Description == "N/A" {						// Could not get description
			atomic.AddUint64(errorCounter, 1)
		}

		wg.Done()
	}

	// Safely increments error counter per SCP
	safeCounter := func(series *[]string, c *uint64) {
		for _, URL := range *series {
			// Concurrently ran in case a specific webpage takes a minute to respond to a request
			wg.Add(1)
			go checkSCP(URL, c)
			// Added to avoid being flagged by the server
			time.Sleep(ServerResponseGuard)
		}
	}

	safeCounter((*allSCPs)[checkSeries], &numErrors)

	wg.Wait()

	return float32(numErrors) / float32(numEntities)
}

// Acquires all series URLs for SCPs
func AcquireAllSCPSeries() *[]string {
	return utils.Map(func(s string) string { return BaseURL + s },
				LinkCollector(BaseURL, SeriesRegexp))
}

// Collects a specified SCPs lab report entry on the SCP-WiKi
func GetSCP(itemNumber int) *scpProcessor.SCPEntity {
	scpSeriesEntry := (itemNumber % 1000)
	if scpSeriesEntry != 0 { scpSeriesEntry-- }
	scpSeries := (itemNumber - scpSeriesEntry) / 1000

	allSeries := AcquireAllSCPSeries()

	seriesSCPs := CollectSCPs(&[]string{(*allSeries)[scpSeries]})

	return SCPInfoCollector((*(*seriesSCPs)[0])[scpSeriesEntry])
}

func GetRange(start, end int) *[]scpProcessor.SCPEntity {
	scps := &[]scpProcessor.SCPEntity{}
	for scpNum := start; scpNum < end; scpNum++ {
		wg.Add(1)
		go func(buildingSCPs *[]scpProcessor.SCPEntity, scpItemNum int) {
			*buildingSCPs = append(*buildingSCPs, *GetSCP(scpItemNum))
			wg.Done()
		}(scps, scpNum)
		time.Sleep(ServerResponseGuard)
	}

	wg.Wait()

	return scps
}