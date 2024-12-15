package crawl

import (
	"context"
	"fmt"
	"github.com/benjaminestes/robots/v2"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/nv4n/go-crawler/fetch/token"
	"github.com/nv4n/go-crawler/model"
	"github.com/nv4n/go-crawler/model/image"
	"github.com/nv4n/go-crawler/utils"
	"log"
	"net/http"
	urlpkg "net/url"
	"time"
)

type Crawler struct {
	PageStore *model.UrlStore
	Browser   *rod.Browser
}

var crawler Crawler

func Close() {
	defer crawler.Browser.MustClose()
	for _, p := range crawler.Browser.MustPages() {
		p.MustClose()
	}
}

func InitCrawler() {
	crawler.PageStore = model.InitUrlStore()
	crawler.Browser = rod.New().MustConnect()

}

func CrawlPage(url string, depth uint, imgChan chan<- image.ImgDownloadInfo, ctx context.Context, rinfo model.RobotsInfo) {
	log.Printf("%s url is being crawled\n", url)
	select {
	case <-ctx.Done():
		<-token.GetReadTokenChan()
		return
	default:
	}

	if !canCrawl(url, depth) {
		utils.Warn("Can't crawl in canCrawl")
		<-token.GetReadTokenChan()

		return
	}
	crawler.PageStore.Add(url)

	if rinfo.RobotsTester == nil {
		r, url := getRobots(url)
		rinfo.RobotsTester = r
		rinfo.URL = url
	}

	if rinfo.URL == "" {
		utils.Warn("Can't crawl no robots URL")
		<-token.GetReadTokenChan()

		return
	}
	if rinfo.URL != "" && rinfo.RobotsTester == nil {
		utils.Warn("Can't crawl no robots tester")

		crawler.PageStore.Add(rinfo.URL)
		<-token.GetReadTokenChan()

		return
	}

	if ok := isRobotsValid(url, rinfo); !ok {
		utils.Warn("Can't crawl  isRobotsValid")

		<-token.GetReadTokenChan()

		return
	}

	if !rinfo.RobotsTester.Test("Go-http-client/1.1", url) {
		utils.Warn("Can't crawl RobotsTester.Test")
		<-token.GetReadTokenChan()

		return
	}
	e := proto.NetworkResponseReceived{}
	page := crawler.Browser.MustPage().MustWindowFullscreen()
	defer page.MustClose()
	wait := page.WaitEvent(&e)
	page.MustNavigate(url)
	wait()

	if e.Response.Status != 200 {
		utils.Warn(fmt.Sprintf("HTTP Error %d: %s", e.Response.Status, e.Response.StatusText))
		<-token.GetReadTokenChan()

		return
	}
	log.Println("Got html page")

	page.MustWaitStable()
	imgs := page.MustElements("img[src]")

	log.Println("Sending images")
	go sendImageData(url, ctx, imgs, imgChan)
	//TODO
	//if *model.ParsedFlags.ExternalLinks {
	//	reader.Find("link[rel=\"stylesheet\"").Each(func(i int, selection *goquery.Selection) {
	//
	//	})
	//
	//}
	go crawlNextPages(page, ctx, depth, imgChan, rinfo)

	<-token.GetReadTokenChan()
}

func crawlNextPages(page *rod.Page, ctx context.Context, depth uint, imgChan chan<- image.ImgDownloadInfo, rinfo model.RobotsInfo) {
	tokenStore := token.GetWriteTokenChan()
	anchors := page.MustElements("a[href]")
	for _, anchor := range anchors {
		select {
		case <-ctx.Done():
			return
		case tokenStore <- struct{}{}:

			href := anchor.MustAttribute("href")
			if href != nil && *href != "" {
				go CrawlPage(*href, depth+1, imgChan, ctx, rinfo)
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func sendImageData(url string, ctx context.Context, imgs rod.Elements, imgChan chan<- image.ImgDownloadInfo) {
	for _, img := range imgs {
		src, err := img.Attribute("src")
		if err != nil {
			*src = ""
		}
		altText, err := img.Attribute("alt")
		if err != nil {
			*altText = "N/A"
		}
		if *src != "" {
			select {
			case <-ctx.Done():
				return
			case imgChan <- image.ImgDownloadInfo{Url: *src, AltText: *altText, RequestUrl: url}:
				log.Printf("Sending image %s\n", *src)
			}
		}
	}
}

func canCrawl(url string, depth uint) bool {
	if crawler.PageStore == nil {
		log.Fatal("URL HTML page store is not initialized")
	}
	if depth > *model.ParsedFlags.DepthLevel {
		return false
	}

	if crawler.PageStore.Contains(url) {
		return false
	}
	return true
}

func isRobotsValid(url string, rinfo model.RobotsInfo) bool {
	parse, err := urlpkg.Parse(url)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR parsing url %s: %+v", url, err))
		return false
	}
	crawlUrlDomain := fmt.Sprintf("%s://%s", parse.Scheme, parse.Host)
	utils.Warn(crawlUrlDomain)
	parse, err = urlpkg.Parse(rinfo.URL)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR parsing url %s: %+v", rinfo.URL, err))
		return false
	}
	robotUrlDomain := fmt.Sprintf("%s://%s", parse.Scheme, parse.Host)

	utils.Warn(fmt.Sprintf("'%s' !== '%s'", crawlUrlDomain, robotUrlDomain))
	return crawlUrlDomain == robotUrlDomain
}

func getRobots(url string) (*robots.Robots, string) {
	robotsUrl, err := robots.Locate(url)
	if err != nil {
		return nil, ""
	}
	resp, err := http.Get(robotsUrl)
	if err != nil {
		return nil, robotsUrl
	}
	defer resp.Body.Close()
	r, err := robots.From(resp.StatusCode, resp.Body)
	if err != nil {
		return nil, robotsUrl
	}
	return r, robotsUrl
}
