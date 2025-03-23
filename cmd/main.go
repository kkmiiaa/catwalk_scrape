package main

import (
	"catwalk_scrape/internal/scraper"
	"context"
	"log"
)

func main() {
	boothScraper := scraper.NewBoothScraper(
		"https://booth.pm/ja/browse/3D%E3%83%A2%E3%83%87%E3%83%AB",
		3027,
		"booth_details.csv",
		"failed_list.csv",
		false,
	)

	err := boothScraper.DoProcess(context.Background())
	if err != nil {
		log.Fatalf("Failed to process: %v", err)
	}
}
