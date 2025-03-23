package scraper

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"

	"catwalk_scrape/internal/utils"
)

type BoothScraper struct {
	BoothTargetUrl  string
	StartPage       int
	EndPage         int
	OutputCSVFile   string
	FailedListFile  string
	AllowAppendLast bool
	AllowRowUpdate  bool // TODO
}

func NewBoothScraper(
	url string,
	startPage int,
	endPage int,
	outputCsvFile string,
	failedListFile string,
	AllowAppendLast bool,
	allowRowUpdate bool,
) *BoothScraper {
	return &BoothScraper{
		BoothTargetUrl:  url,
		StartPage:       startPage,
		EndPage:         endPage,
		OutputCSVFile:   outputCsvFile,
		FailedListFile:  failedListFile,
		AllowAppendLast: AllowAppendLast,
		AllowRowUpdate:  allowRowUpdate,
	}
}

func (bs *BoothScraper) DoProcess(
	ctx context.Context,
) error {
	log.Printf("Start to process url: %s\n", bs.BoothTargetUrl)

	// scrape summary page
	summaryItems, err := bs.scrapeBoothList(ctx, bs.BoothTargetUrl)
	if err != nil {
		log.Fatalf("Failed to scrape summary page: %s, err: %v", bs.BoothTargetUrl, err)
		return err
	}
	log.Println("Succesfully scraped summary page, item count: ", len(summaryItems))

	// init output file
	file, writer, err := bs.initOutputFile(ctx, bs.OutputCSVFile)
	if err != nil {
		log.Fatalf("Failed to init output file, err: %v", err)
		return err
	}
	defer file.Close()
	defer writer.Flush()

	failedFile, failedWriter, err := bs.initFailedFile(ctx, bs.FailedListFile)
	if err != nil {
		log.Fatalf("Failed to init output file, err: %v", err)
		return err
	}
	defer failedFile.Close()
	defer failedWriter.Flush()

	// scrape detail page
	itemCount, _ := bs.scrapeAndSaveBoothItemDetails(ctx, summaryItems, writer, failedWriter)
	log.Println("Succesfully scraped detail page, item count: ", itemCount)

	return nil
}

func (bs *BoothScraper) scrapeBoothList(
	ctx context.Context,
	baseUrl string,
) ([]BoothSummaryItem, error) {
	c := colly.NewCollector()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	processChan := make(chan struct{}, 5)

	var results []BoothSummaryItem
	var mu sync.Mutex
	var wg sync.WaitGroup

	for pageNum := bs.StartPage; pageNum <= bs.EndPage; pageNum++ {
		<-ticker.C
		processChan <- struct{}{}

		wg.Add(1)
		go func() {
			defer func() { <-processChan }()
			fetchItemsFromBoothSearchResult(ctx, c, baseUrl, pageNum, &results, &mu, &wg)
		}()
	}
	wg.Wait()

	return results, nil
}

func fetchItemsFromBoothSearchResult(
	ctx context.Context,
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
				Title:   e.Text,
				Url:     e.Attr("href"),
				PageNum: pageNum,
			}

			results = append(results, item)
		},
	)

	url := baseUrl + "?page=" + strconv.Itoa(pageNum)

	log.Println("start to scrape page: ", url)
	c.Visit(url)

	log.Println("Successfully finished scraping: ", url)

	mu.Lock()
	*parentResults = append(*parentResults, results...)
	mu.Unlock()
}

func (bs *BoothScraper) scrapeAndSaveBoothItemDetails(
	ctx context.Context,
	summaryItems []BoothSummaryItem,
	writer *csv.Writer,
	failedWriter *csv.Writer,
) (int, error) {
	ticker := time.NewTicker(330 * time.Millisecond)
	defer ticker.Stop()

	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Headless)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	processChan := make(chan struct{}, 10)

	itemChannel := make(chan BoothDetailItem, 100)
	var csvWg sync.WaitGroup
	csvWg.Add(1)
	go bs.saveRecordToCSV(writer, itemChannel, &csvWg)

	var fetchWg sync.WaitGroup

	for i, item := range summaryItems {
		<-ticker.C
		processChan <- struct{}{}
		fetchWg.Add(1)

		go func(idx int, summaryItem BoothSummaryItem) {
			defer func() { <-processChan }()
			bs.fetchItemDetail(allocCtx, summaryItem, &fetchWg, itemChannel, failedWriter)
		}(i, item)
	}

	fetchWg.Wait()
	close(itemChannel)
	csvWg.Wait()

	return len(summaryItems), nil
}

func (bs *BoothScraper) fetchItemDetail(
	ctx context.Context,
	summaryItem BoothSummaryItem,
	wg *sync.WaitGroup,
	itemChannel chan<- BoothDetailItem,
	failedWriter *csv.Writer,
) {
	defer wg.Done()

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var wholeHtml string
	err := chromedp.Run(
		ctx,
		chromedp.Navigate(summaryItem.Url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.OuterHTML("html", &wholeHtml, chromedp.ByQuery),
	)
	if err != nil {
		log.Printf("Failed to run chromedp: %v", err)
		if err := failedWriter.Write([]string{summaryItem.Url}); err != nil {
			log.Println("=========== Failed to save record to csv. url: ", summaryItem.Url)
			return
		}
		return
	}

	dom, err := goquery.NewDocumentFromReader(strings.NewReader(wholeHtml))
	if err != nil {
		log.Printf("Failed to run chromedp: %v", err)
		if err := failedWriter.Write([]string{summaryItem.Url}); err != nil {
			log.Println("=========== Failed to save record to csv. url: ", summaryItem.Url)
			return
		}
		return
	}

	var item BoothDetailItem

	jst, _ := time.LoadLocation("Asia/Tokyo")

	item.Title = dom.Find("header > h2").Text()
	item.Url = summaryItem.Url

	item.CreatorName = dom.Find(".my-16 header div.flex a.grid span").Text()
	item.CreatorUrl = dom.Find(".my-16 header div.flex a.grid").AttrOr("href", "")

	var priceOptions []BoothItemPriceOption
	dom.Find(".variation-item").Each(
		func(index int, s *goquery.Selection) {
			title := s.Find(".variation-name").Text()

			jpyPriceStr := s.Find(".variation-price").Text()
			jpyPriceStr = utils.TrimString(jpyPriceStr, "¥", "")

			priceOption := BoothItemPriceOption{
				Title:    title,
				JpyPrice: jpyPriceStr,
			}
			priceOptions = append(priceOptions, priceOption)
		},
	)
	item.PriceOptions = priceOptions

	publishedAtStr := dom.Find("#js-item-published-date .typography-14").Text()
	publishedAtStr = utils.TrimString(publishedAtStr, "商品公開日時：", "")
	item.PublishedAt, _ = time.ParseInLocation("2006年1月2日 15時04分", publishedAtStr, jst)

	salesStartAtStr := dom.Find("div.my-16 > header > .typography-14").Text()
	salesStartAtStr = utils.TrimString(salesStartAtStr, "販売期間：", "から")
	item.SalesStartAt, _ = time.ParseInLocation("2006年1月2日 15時04分", salesStartAtStr, jst)

	item.Description = dom.Find("section.main-info-column").Text() + "\n" + dom.Find("div.my-40").Text()
	item.DescriptionWordCount = utf8.RuneCountInString(item.Description)

	// Tags
	var tags []string
	dom.Find("#js-item-tag-list div.gap-8 > a.no-underline").Each(
		func(index int, s *goquery.Selection) {
			tags = append(tags, s.Text())
		},
	)
	item.Tags = tags

	item.FavoriteCount, _ = strconv.Atoi(dom.Find("#js-item-wishlist-button .typography-14").Text())
	item.PictureCount = dom.Find(".slick-list > .slick-track > .slick-slide:not(.slick-cloned) > div > a > div > img").Length()

	item.ScrapeAt = time.Now()
	item.PageNum = summaryItem.PageNum

	log.Println("Successfully finished scraping detail: ", item.Url)
	itemChannel <- item
}

func (bs *BoothScraper) initOutputFile(
	ctx context.Context,
	filename string,
) (*os.File, *csv.Writer, error) {
	path := "output/" + filename
	var file *os.File
	var err error

	if bs.AllowAppendLast {
		file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		file, err = os.Create(path)
	}
	if err != nil {
		return nil, nil, err
	}
	writer := csv.NewWriter(file)

	if !bs.AllowAppendLast {
		header := []string{
			"Title", "Url", "CreatorName", "CreatorUrl", "JpyPrice",
			"PublishedAt", "SalesStartAt", "Description", "DescriptionWordCount",
			"Tags", "FavoriteCount", "PictureCount", "PageNum", "ScrapeAt",
		}

		if err := writer.Write(header); err != nil {
			return nil, nil, err
		}
	}

	return file, writer, nil
}

func (bs *BoothScraper) initFailedFile(
	ctx context.Context,
	filename string,
) (*os.File, *csv.Writer, error) {
	file, err := os.Create("output/" + filename)
	if err != nil {
		return nil, nil, err
	}
	writer := csv.NewWriter(file)

	header := []string{"pageUrl"}
	if err := writer.Write(header); err != nil {
		return nil, nil, err
	}

	return file, writer, nil
}

func (bs *BoothScraper) saveRecordToCSV(
	writer *csv.Writer,
	itemChannel <-chan BoothDetailItem,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for item := range itemChannel {
		priceOptionsJson, err := json.Marshal(item.PriceOptions)
		if err != nil {
			log.Println("Failed to marshal PriceOptions:", err)
			priceOptionsJson = []byte("[]") // 空の配列として出力
		}

		tagsJson, err := json.Marshal(item.Tags)
		if err != nil {
			log.Println("Failed to marshal Tags:", err)
			tagsJson = []byte("[]")
		}

		record := []string{
			item.Title,
			item.Url,
			item.CreatorName,
			item.CreatorUrl,
			string(priceOptionsJson),
			item.PublishedAt.Format(time.RFC3339),  // ISO 8601 形式
			item.SalesStartAt.Format(time.RFC3339), // ISO 8601 形式
			item.Description,
			strconv.Itoa(item.DescriptionWordCount),
			string(tagsJson),
			strconv.Itoa(item.FavoriteCount),
			strconv.Itoa(item.PictureCount),
			strconv.Itoa(item.PageNum),
			item.ScrapeAt.Format(time.RFC3339),
		}

		if err := writer.Write(record); err != nil {
			log.Println("Failed to save record to csv. title: ", item.Title)
			return
		}
	}
}
