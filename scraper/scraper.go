package scraper

import (
	"reflect"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gocolly/colly/v2"
	"github.com/kataras/iris/v12"

	"pricetracker/models"
)

var Scraper *colly.Collector

func Init() {
	Scraper = colly.NewCollector()

    s := gocron.NewScheduler(time.UTC)
    _, err := s.Every(24).Hours().Do(Scrape)
    if err != nil {
        panic(err)
    }
}

func Scrape() {
    ctxPtr := reflect.New(reflect.TypeOf(new(iris.Context)))
    ctx := ctxPtr.Elem().Interface().(iris.Context)

    trackers := models.GetActiveTrackers(ctx)
    for _, tracker := range trackers {
        item := models.GetItem(tracker.ItemId, ctx)
        for _, merchant := range item.Merchants {
            price := models.Price{}
            price.MerchantId = merchant.Id

            if (merchant.Name == "Amazon") {
                p := ScrapeAmazon(merchant.Url)
                price.Value = &p
            }
            if (merchant.Name == "Wayfair") {
                p := ScrapeWayfair(merchant.Url)
                price.Value = &p
            }

            models.CreatePrice(&price, ctx)
        }
    }

}

func ScrapeAmazon(url string) float64 {
    var price float64

    Scraper.OnHTML(".SFPrice span", func(e *colly.HTMLElement) {
        flt, err := strconv.ParseFloat(e.Text, 64)
        if err != nil {
            panic(err)
        }
        price = flt
    })

    Scraper.Visit(url)
    return price
}

func ScrapeWayfair(url string) float64 {
    var price float64

    Scraper.OnHTML(".a-price-whole", func(e *colly.HTMLElement) {
        flt, err := strconv.ParseFloat(e.Text, 64)
        if err != nil {
            panic(err)
        }
        price = flt
    })

    Scraper.OnHTML(".a-price-fraction", func(e *colly.HTMLElement) {
        flt, err := strconv.ParseFloat("0." + e.Text, 64)
        if err != nil {
            panic(err)
        }
        price += flt
    })

    Scraper.Visit(url)
    return price
}
