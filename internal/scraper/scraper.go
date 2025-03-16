package scraper

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

var rules = []ScrapeRule{
	{
		regexp.MustCompile(`booth\.pm/ja/browse/.*`),
		ScrapeBoothSearchResult,
	},
	{
		regexp.MustCompile(`booth\.pm/ja/items/\d+`),
		ScrapeBoothItemDetail,
	},
}

func Scrape(
	url string,
) ([]string, error) {
	for _, rule := range rules {
		if rule.Pattern.MatchString(url) {
			return rule.Handler(colly.NewCollector(), url)
		}
	}
	return nil, errors.New("scraping rule not exist")
}

func ScrapeBoothSearchResult(
	c *colly.Collector,
	baseUrl string,
) ([]string, error) {
	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	var results []BoothSummaryItem
	var mu sync.Mutex
	var wg sync.WaitGroup

	pageNum := 0
	for range ticker.C {
		pageNum++

		wg.Add(1)
		go fetchItemsFromBoothSearchResult(c, baseUrl, pageNum, &results, &mu, &wg)

		if pageNum >= 10000 {
			break
		}
	}

	wg.Wait()
	log.Println("scraperd booth search result", results[:100])

	return []string{}, nil
}

func fetchItemsFromBoothSearchResult(
	c *colly.Collector,
	baseUrl string,
	pageNum int,
	parentResults *[]BoothSummaryItem,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	var results []BoothSummaryItem

	c.OnHTML(
		"li.item-card .item-card__wrap .item-card__summary .item-card__title a",
		func(e *colly.HTMLElement) {
			item := BoothSummaryItem{
				Title: e.Text,
				Url:   e.Attr("href"),
			}

			results = append(results, item)
		},
	)

	url := baseUrl + "?page=" + strconv.Itoa(pageNum)

	log.Println("start to scrape page: ", url)
	c.Visit(url)

	log.Println("Successfully finished scraping: ", results)

	mu.Lock()
	*parentResults = append(*parentResults, results...)
	mu.Unlock()
}

func ScrapeBoothItemDetail(
	c *colly.Collector,
	url string,
) ([]string, error) {
	var results []string
	c.OnHTML(".headline, .summary", func(e *colly.HTMLElement) {
		results = append(results, e.Text)
	})
	err := c.Visit(url)
	return results, err
}
