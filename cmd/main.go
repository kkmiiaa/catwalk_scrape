package main

import (
	"catwalk_scrape/internal/scraper"
	"log"
)

func main() {
	url := "https://example.com"

	data, err := scraper.Scrape(url)
	if err != nil {
		log.Fatalf("failed to scrape: %v", err)
	}

	log.Println("succesfully", data)
}
