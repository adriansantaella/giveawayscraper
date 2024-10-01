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

	"github.com/PuerkitoBio/goquery"
)

var wg sync.WaitGroup
var eligible = []string{"worldwide", "us", "north america", "everywhere"}

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
	defer func() {
		items = items[:0] // Clear the slice when the function exits
	}()

	for i := 1; i <= numofpages; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("Scraping page %d\n from %s\n", i, url)
			pageUrl := url + strconv.Itoa(i)

			resp, err := http.Get(pageUrl)
			if err != nil {
				fmt.Printf("Error visiting page %d: %v\n", i, err)
				return
			}
			defer resp.Body.Close()

			fmt.Println(resp.StatusCode)

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("HTTP request failed with status: %d\n", resp.StatusCode)
				return
			}

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return
			}

			processPage(doc)
		}(i)
	}

	wg.Wait()

	return items, nil
}

func processPage(doc *goquery.Document) {
	var pageItems []Item

	links := doc.Find("a.read-more")

	links.Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		wg.Add(1)
		go func() {
			defer wg.Done()
			visitLink(href, &pageItems)
		}()
	})

	items = append(items, pageItems...)
}

func visitLink(href string, pageItems *[]Item) {
	linkDoc, err := goquery.NewDocument(href)
	if err != nil {
		fmt.Printf("Error fetching link: %v\n", err)
		return
	}

	linkDoc.Find(".inside-article").Each(func(i int, article *goquery.Selection) {
		isEligible := false
		isExpired := true
		link := ""
		expires := ""
		imageUrl, _ := article.Find(".attachment-full").Attr("src")

		entryTitle := article.Find(".entry-title").Text()
		elements := article.Find("p, h3")

		elements.Each(func(_ int, el *goquery.Selection) {
			if strings.Contains(el.Text(), "OPEN TO:") {
				for _, word := range eligible {
					if strings.Contains(strings.ToLower(el.Text()), word) {
						isEligible = true
					}
				}
			}

			if strings.Contains(el.Text(), "GIVEAWAY END") {
				expires = el.Text()[14:]
				diff := parseAndCompareDate(expires)
				if diff >= 0 {
					isExpired = false
				}
			}

			if strings.Contains(el.Text(), "STEP 1") {
				child := el.Find("span a")
				linkAttr, exists := child.Attr("href")
				if !exists {
					fmt.Println("href attribute not found")
				} else {
					link = linkAttr
				}
			}
		})

		if isEligible && !isExpired {
			item := Item{
				ExpirationDate: expires,
				Name:           entryTitle,
				URL:            link,
				ImageURL:       imageUrl,
			}

			*pageItems = append(*pageItems, item)
		}

	})
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
