package main

import (
	"catwalk_scrape/internal/scraper"
	"log"
)

func main() {
	url := "https://booth.pm/ja/browse/3D%E3%83%A2%E3%83%87%E3%83%AB"

	data, err := scraper.Scrape(url)
	if err != nil {
		log.Fatalf("failed to scrape: %v", err)
	}

	log.Println("succesfully", data)
}
