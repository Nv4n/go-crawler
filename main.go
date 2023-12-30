package main

import (
	"context"
	"flag"
	_ "github.com/PuerkitoBio/goquery"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/go-playground/validator/v10"
	"log"
	"time"
)

type ImageData struct {
	Filename   string `db:"filename" json:"filename"`
	AltText    string `db:"alt_text" json:"alt_text"`
	Title      string `db:"title" json:"title"`
	Resolution string `db:"resolution" json:"resolution"`
	Format     string `db:"format" json:"format"`
}

type Flags struct {
	Url           *string `validate:"url"`
	ExternalLinks *bool
	Spa           *bool
	Timeout       *int  `validate:"min=1,max=10"`
	Goroutines    *uint `validate:"min=5,max=50"`
	DepthLevel    *uint `validate:"min=1,max=10"`
}

var flags Flags
var tokenStore chan struct{}

var validate *validator.Validate

func init() {
	flags = Flags{}
	validate = validator.New()

	flags.Url = flag.String("url", "", "URL to be web-crawled for images")
	flags.Spa = flag.Bool("spa", false, "Is the site SPA (client-rendered)")
	flags.ExternalLinks = flag.Bool("el", false, "Follow external links")
	flags.DepthLevel = flag.Uint("dl", 3, "Depth level of image crawling")
	flags.Timeout = flag.Int("t", 2, "Minutes before timeout the execution")
	flags.Goroutines = flag.Uint("g", 20, "Maximum goroutines")
}
func setupCrawler() (context.Context, context.CancelFunc) {
	flag.PrintDefaults()
	flag.Parse()
	err := validate.Struct(flags)

	if err != nil {
		log.Fatalf("Validation errors: %+v", err)
	}
	tokenStore = make(chan struct{}, *flags.Goroutines)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*flags.Timeout)*time.Minute)
	return ctx, cancel
}

func main() {
	ctx, cancel := setupCrawler()
	defer close(tokenStore)
	defer cancel()
}
