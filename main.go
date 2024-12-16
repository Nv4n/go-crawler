package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/nv4n/go-crawler/fetch/crawl"
	"github.com/nv4n/go-crawler/fetch/db"
	"github.com/nv4n/go-crawler/fetch/img"
	"github.com/nv4n/go-crawler/fetch/token"
	"github.com/nv4n/go-crawler/model"
	"github.com/nv4n/go-crawler/model/image"
	"github.com/nv4n/go-crawler/views"
	"log"
	"net/http"
	"time"
)

var validate *validator.Validate
var imgDownloadDataChan chan image.ImgDownloadInfo

func init() {
	model.ParsedFlags = model.CliFlags{}
	validate = validator.New()
	model.ParsedFlags.Url = flag.String("url", "", "URL to be web-crawled for images")
	model.ParsedFlags.Spa = flag.Bool("spa", false, "Is the site SPA (client-rendered)")
	model.ParsedFlags.ExternalLinks = flag.Bool("el", false, "Follow external links")
	model.ParsedFlags.DepthLevel = flag.Uint("dl", 5, "Depth level of image crawling")
	model.ParsedFlags.Timeout = flag.Int("t", 2, "Minutes before timeout the execution")
	model.ParsedFlags.Goroutines = flag.Uint("g", 20, "Maximum goroutines")
	model.ParsedFlags.AcceptElement = flag.String("cookieel", "button", "Element to click and accept cookies")
	model.ParsedFlags.AcceptTxt = flag.String("cookietxt", "Accept All Cookies", "Text inside accept cookies element")
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
	crawl.InitCrawler()
	db.InitDb()
	imgDownloadDataChan = make(chan image.ImgDownloadInfo)

	defer crawl.Close()
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
	http.HandleFunc("/filter", handleImageFilter)
	fmt.Println("Listening to :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleImagePage(w http.ResponseWriter, r *http.Request) {
	images := db.GetAllImages()
	set := make(map[int]struct{})
	imageChan := make(chan image.DbMetadata)
	ctx := r.Context()
	go func(ctxGoroutine context.Context) {
		defer close(imageChan)
		var ids []int32
		for _, imgData := range images {
			select {
			case <-ctxGoroutine.Done():
				return
			default:
				if _, ok := set[imgData.Id]; !ok {
					ids = append(ids, int32(imgData.Id))
					set[imgData.Id] = struct{}{}
					select {
					case <-ctxGoroutine.Done():
						return
					case imageChan <- imgData:
					}
				}
			}
		}

		for {
			images := db.GetAllImagesWithoutIds(ids)
			for _, imgData := range images {
				select {
				case <-ctxGoroutine.Done():
					return
				default:
					if _, ok := set[imgData.Id]; !ok {
						ids = append(ids, int32(imgData.Id))
						set[imgData.Id] = struct{}{}
						select {
						case <-ctxGoroutine.Done():
							return
						case imageChan <- imgData:
						}
					}
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

	}(ctx)
	templ.Handler(views.Page(imageChan), templ.WithStreaming()).ServeHTTP(w, r)
}

func handleImageFilter(w http.ResponseWriter, r *http.Request) {
	filter := image.DbFilter{
		Title:   r.URL.Query().Get("title"),
		AltText: r.URL.Query().Get("alt_text"),
		Format:  r.URL.Query().Get("format"),
	}
	log.Println(fmt.Sprintf("FILTER IS HERE: %v", filter))

	images := db.GetFilteredImages(filter)
	set := make(map[int]struct{})
	imageChan := make(chan image.DbMetadata)
	ctx := r.Context()
	go func(ctxGoroutine context.Context) {
		defer close(imageChan)
		for _, imgData := range images {
			select {
			case <-ctxGoroutine.Done():
				return
			default:
				if _, ok := set[imgData.Id]; !ok {
					set[imgData.Id] = struct{}{}
					select {
					case <-ctxGoroutine.Done():
						return
					case imageChan <- imgData:
					}
				}
			}
		}
	}(ctx)
	templ.Handler(views.Page(imageChan), templ.WithStreaming()).ServeHTTP(w, r)
}
