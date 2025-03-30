package main

import (
	"catwalk_scrape/internal/scraper"
	"context"
	"log"
)

func main() {
	boothScraper := scraper.NewBoothScraper(
		&scraper.BoothScraperOptions{
			BoothTargetUrl:  "https://booth.pm/ja/browse/3D%E3%83%A2%E3%83%87%E3%83%AB",
			StartPage:       1,
			EndPage:         10,
			OutputCSVFile:   "booth_items.csv",
			FailedListFile:  "failed_list.csv",
			AllowAppendLast: true,
			AllowRowUpdate:  false, // NOTE: 未実装
		},
	)

	err := boothScraper.DoProcess(context.Background())
	if err != nil {
		log.Fatalf("Failed to process: %v", err)
	}
}
