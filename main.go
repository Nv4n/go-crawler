package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/nv4n/go-crawler/fetch/crawl"
	"github.com/nv4n/go-crawler/fetch/db"
	"github.com/nv4n/go-crawler/fetch/img"
	"github.com/nv4n/go-crawler/fetch/token"
	"github.com/nv4n/go-crawler/model"
	"github.com/nv4n/go-crawler/model/image"
	"html/template"
	"log"
	"net/http"
	"time"
)

var validate *validator.Validate

func init() {
	model.ParsedFlags = model.CliFlags{}
	validate = validator.New()

	model.ParsedFlags.Url = flag.String("url", "", "URL to be web-crawled for images")
	model.ParsedFlags.Spa = flag.Bool("spa", false, "Is the site SPA (client-rendered)")
	model.ParsedFlags.ExternalLinks = flag.Bool("el", false, "Follow external links")
	model.ParsedFlags.DepthLevel = flag.Uint("dl", 3, "Depth level of image crawling")
	model.ParsedFlags.Timeout = flag.Int("t", 2, "Minutes before timeout the execution")
	model.ParsedFlags.Goroutines = flag.Uint("g", 20, "Maximum goroutines")
}
func setupCrawler() (context.Context, context.CancelFunc) {
	flag.PrintDefaults()
	flag.Parse()
	err := validate.Struct(model.ParsedFlags)

	if err != nil {
		log.Fatalf("Validation errors: %+v", err)
	}
	token.InitTokenStore(*model.ParsedFlags.Goroutines)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*model.ParsedFlags.Timeout)*time.Minute)
	return ctx, cancel
}

func main() {
	ctx, cancel := setupCrawler()
	defer cancel()

	token.InitTokenStore(*model.ParsedFlags.Goroutines)
	img.InitImageStore()
	crawl.InitPageStore()
	db.InitDb()
	imgDownloadDataChan := make(chan image.ImgDownloadInfo)

	defer token.Close()
	defer db.CloseDb()
	defer close(imgDownloadDataChan)

	tokenStoreSend := token.GetWriteTokenChan()
	tokenStoreSend <- struct{}{}
	tokenStoreSend <- struct{}{}
	go crawl.CrawlPage(*model.ParsedFlags.Url, 1, imgDownloadDataChan, ctx, model.RobotsInfo{})
	go img.FetchImages(imgDownloadDataChan, ctx, tokenStoreSend)

	staticFs := http.FileServer(http.Dir("./static"))
	uploadsFs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", uploadsFs))
	http.Handle("/static/", http.StripPrefix("/static/", staticFs))
	http.HandleFunc("/", handleImagePage)
	fmt.Println("Listening to :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleImagePage(w http.ResponseWriter, _ *http.Request) {
	tmpls := template.Must(template.ParseFiles("views/index.go.html", "views/image.go.html"))
	err := tmpls.ExecuteTemplate(w, "Base", db.GetAllImages())
	if err != nil {
		w.WriteHeader(500)
	}
}
