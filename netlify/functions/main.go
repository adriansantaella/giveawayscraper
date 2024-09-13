package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/adriansantaella/giveawayscraper/scraper"
)

type APIResponse struct {
	Items []scraper.Item `json:"items"`
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	numofpages := r.URL.Query().Get("numpages")

	numPages, err := strconv.Atoi(numofpages)

	if err != nil {
		fmt.Printf("Error converting number of pages passed in: %v\n", err)
		return
	}

	items, err := scraper.ScrapeData("https://www.giveawaybase.com/page/", numPages)
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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	w.Write(jsonResp)

}

func main() {
	http.HandleFunc("/scrape-data", handleRequest)
	http.ListenAndServe(":8080", nil)
}
