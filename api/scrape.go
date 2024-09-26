package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

type Item struct {
	ExpirationDate string
	Name           string
	URL            string
	ImageURL       string
}

type APIItem interface {
	GetName() string
	GetURL() string
	GetExpirationDate() string
}

var items []Item

type APIResponse struct {
	Items []Item `json:"items"`
}

func Scrape(w http.ResponseWriter, r *http.Request) {
	// Handle preflight CORS request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set CORS headers for actual request
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	numofpages := r.URL.Query().Get("numpages")

	numPages, err := strconv.Atoi(numofpages)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error converting number of pages passed in: %v", err), http.StatusBadRequest)
		return
	}

	items, err := ScrapeData("https://www.giveawaybase.com/page/", numPages)

	if err != nil {
		http.Error(w, fmt.Sprintf("Scraping error: %v", err), http.StatusInternalServerError)
		return
	}

	apiResponse := APIResponse{
		Items: items,
	}

	jsonResp, err := json.Marshal(apiResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(jsonResp); err != nil {
		http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
	}

}

func ScrapeData(url string, numofpages int) ([]Item, error) {
	fmt.Println("Called Scrape Data")

	defer func() {
		items = items[:0] // Clear the slice when the function exits
	}()

	c := colly.NewCollector(
		colly.Async(true),
	)

	eligible := []string{"worldwide", "us", "north america", "everywhere"}

	var wg sync.WaitGroup

	c.OnHTML("a.read-more", func(e *colly.HTMLElement) {
		// grab the first READ MORE link from each page
		link := e.Attr("href")
		c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnHTML(".inside-article", func(e *colly.HTMLElement) {
		// Loop over each child element with the class .child
		isEligible := false
		isExpired := true
		link := ""
		expires := ""
		imageUrl := e.ChildAttr(".attachment-full", "src")

		entryTitle := e.ChildText(".entry-title")
		fmt.Printf("Title: %s\n", entryTitle)

		e.ForEach("p, h3", func(_ int, el *colly.HTMLElement) {

			// Check value of eligibility
			if strings.Contains(el.Text, "OPEN TO:") {
				for _, word := range eligible {
					if strings.Contains(strings.ToLower(el.Text), word) {
						isEligible = true
					}
				}
			}

			if strings.Contains(el.Text, "GIVEAWAY END") {
				expires = el.Text[14:]
				diff := parseAndCompareDate(expires)
				if diff >= 0 {
					isExpired = false
				}
			}

			if strings.Contains(el.Text, "STEP 1") {
				link = el.ChildAttr("span a", "href")
			}

		})

		if isEligible && !isExpired {
			items = append(items, Item{
				ExpirationDate: expires,
				Name:           entryTitle,
				URL:            link,
				ImageURL:       imageUrl,
			})
		}

	})

	for i := 1; i <= numofpages; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("Scraping page %d\n from %s\n", i, url)
			err := c.Visit(url + strconv.Itoa(i))
			if err != nil {
				fmt.Printf("Error visiting page %d: %v\n", i, err)
			}
		}(i)

	}

	wg.Wait()
	c.Wait()

	fmt.Printf("Number of items collected: %v\n", items)

	return items, nil
}

func parseAndCompareDate(inputString string) int64 {
	// Define th layout based on your input string format
	layout := "January 2, 2006"

	// Remove the suffixes (st, nd, rd, th)
	re := regexp.MustCompile(`(\d+)(st|nd|rd|th)`)
	cleanedDateString := re.ReplaceAllString(inputString, `$1`)

	// Parse the input string into a time.Time
	t, err := time.Parse(layout, cleanedDateString)
	if err != nil {
		fmt.Printf("Error parsing date: %v\n", err)
		return -1 // Return an invalid value to indicate error
	}

	now := time.Now()
	diff := t.Sub(now)
	millisecondsDiff := int64(diff.Milliseconds())

	return millisecondsDiff
}
