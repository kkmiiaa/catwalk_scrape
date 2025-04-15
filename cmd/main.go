package main

import (
	"catwalk_scrape/internal/scraper"
	"context"
	"log"
)

func main() {
	boothScraper := scraper.NewBoothScraper(
		&scraper.BoothScraperOptions{
			BoothTargetUrl:  "https://booth.pm/ja/browse/3D%E3%82%AD%E3%83%A3%E3%83%A9%E3%82%AF%E3%82%BF%E3%83%BC?q=VRChat",
			StartPage:       1,
			EndPage:         1,
			OutputCSVFile:   "booth_items.csv",
			FailedListFile:  "failed_list.csv",
			AllowAppendLast: false,
			AllowRowUpdate:  false, // NOTE: 未実装
		},
	)

	err := boothScraper.DoProcess(context.Background())
	if err != nil {
		log.Fatalf("Failed to process: %v", err)
	}
}
