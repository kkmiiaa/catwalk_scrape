package scraper

import (
	"catwalk_scrape/internal/utils"
	"fmt"

	"github.com/gocolly/colly"
)

// Scrape は指定したURLのデータを取得する関数
func Scrape(url string) ([]string, error) {
	var results []string

	c := colly.NewCollector()

	c.OnHTML("h1, h2, h3", func(e *colly.HTMLElement) {
		results = append(results, e.Text)
	})

	c.OnError(func(r *colly.Response, err error) {
		utils.LogError(fmt.Sprintf("Error: %s", err))
	})

	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	return results, nil
}
