package main

import (
	"context"
	"flag"
	_ "github.com/PuerkitoBio/goquery"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/go-playground/validator/v10"
	"github.com/nv4n/go-crawler/fetch/crawl"
	"github.com/nv4n/go-crawler/fetch/db"
	"github.com/nv4n/go-crawler/fetch/img"
	"github.com/nv4n/go-crawler/fetch/token"
	"github.com/nv4n/go-crawler/model"
	"log"
	"net/http"
	"time"
)

var parsedFlags model.CliFlags
var validate *validator.Validate

func init() {
	parsedFlags = model.CliFlags{}
	validate = validator.New()

	parsedFlags.Url = flag.String("url", "", "URL to be web-crawled for images")
	parsedFlags.Spa = flag.Bool("spa", false, "Is the site SPA (client-rendered)")
	parsedFlags.ExternalLinks = flag.Bool("el", false, "Follow external links")
	parsedFlags.DepthLevel = flag.Uint("dl", 3, "Depth level of image crawling")
	parsedFlags.Timeout = flag.Int("t", 2, "Minutes before timeout the execution")
	parsedFlags.Goroutines = flag.Uint("g", 20, "Maximum goroutines")
}
func setupCrawler() (context.Context, context.CancelFunc) {
	flag.PrintDefaults()
	flag.Parse()
	err := validate.Struct(parsedFlags)

	if err != nil {
		log.Fatalf("Validation errors: %+v", err)
	}
	token.InitTokenStore(*parsedFlags.Goroutines)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*parsedFlags.Timeout)*time.Minute)
	return ctx, cancel
}

func main() {
	ctx, cancel := setupCrawler()
	defer close(tokenStore)
	defer cancel()
	defer db.Db.Close()

	img.InitImageStore()
	crawl.InitPageStore()

	imgUrlChan := make(chan struct{})

	for {
		select {
		case <-ctx.Done():
			log.Fatal(http.ListenAndServe(":8080", nil))
		}

	}

}
