package scraper

import (
	"regexp"

	"github.com/gocolly/colly"
)

type PageHandler func(*colly.Collector, string) ([]string, error)

type ScrapeRule struct {
	Pattern *regexp.Regexp
	Handler PageHandler
}

type BoothSummaryResultPage struct {
	PageNum int
	Url     string
}

type BoothSummaryItem struct {
	Title string
	Url   string
}
