package scraper

import (
	"regexp"
	"time"
)

type PageHandler func(string) ([]string, error)

type ScrapeRule struct {
	Pattern *regexp.Regexp
	Handler PageHandler
}

type BoothSummaryResultPage struct {
	PageNum int
	Url     string
}

type BoothSummaryItem struct {
	Title   string
	Url     string
	PageNum int
}

type BoothDetailItem struct {
	Title string
	Url   string

	CreatorName string
	CreatorUrl  string

	PriceOptions []BoothItemPriceOption

	PublishedAt  time.Time
	SalesStartAt time.Time

	Description          string
	DescriptionWordCount int

	Tags []string

	FavoriteCount int
	PictureCount  int

	ScrapeAt time.Time
	PageNum  int
}

type BoothItemPriceOption struct {
	Title    string
	JpyPrice string
}
