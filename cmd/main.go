package main

import (
	"catwalk_scrape/internal/scraper"
	"context"
	"log"
)

func main() {
	boothScraper := scraper.NewBoothScraper(
		"https://booth.pm/ja/browse/3D%E3%83%A2%E3%83%87%E3%83%AB",
		2,  // start page num
		10, // end page num
		"booth_items.csv",
		"failed_list.csv",
		true,
		false,
	)

	err := boothScraper.DoProcess(context.Background())
	if err != nil {
		log.Fatalf("Failed to process: %v", err)
	}
}
