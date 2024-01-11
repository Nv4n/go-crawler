package crawl

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/benjaminestes/robots/v2"
	"github.com/nv4n/go-crawler/fetch/token"
	"github.com/nv4n/go-crawler/model"
	"github.com/nv4n/go-crawler/model/image"
	"github.com/nv4n/go-crawler/utils"
	"log"
	"net/http"
	urlpkg "net/url"
	"time"
)

var pageStore *model.UrlStore

func InitPageStore() {
	pageStore = model.InitUrlStore()
}

func CrawlPage(url string, depth uint, imgChan chan<- image.ImgDownloadInfo, ctx context.Context, rinfo model.RobotsInfo) {
	log.Printf("%s url is being crawled\n", url)
	select {
	case <-ctx.Done():
		return
	default:
	}

	if !canCrawl(url, depth) {
		return
	}
	pageStore.Add(url)

	if rinfo.RobotsTester == nil {
		r, url := getRobots(url)
		rinfo.RobotsTester = r
		rinfo.URL = url
	}

	if rinfo.URL == "" {
		return
	}
	if rinfo.URL != "" && rinfo.RobotsTester == nil {
		pageStore.Add(rinfo.URL)
		return
	}

	if ok := isRobotsValid(url, rinfo); !ok {
		return
	}

	if !rinfo.RobotsTester.Test("Go-http-client/1.1", url) {
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR fetching HTML for %s: %+v", url, err))
		return
	}
	if resp.StatusCode != 200 {
		utils.Warn(fmt.Sprintf("HTTP Error %d: %s", resp.StatusCode, resp.Status))
		return
	}
	log.Println("Got html page")

	defer resp.Body.Close()
	reader, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR parsing HTML document: %+v", err))
		return
	}
	log.Println("Sending images")
	go sendImageData(url, ctx, reader, imgChan)
	//if *model.ParsedFlags.ExternalLinks {
	//	reader.Find("link[rel=\"stylesheet\"").Each(func(i int, selection *goquery.Selection) {
	//
	//	})
	//
	//}
	go crawlNextPages(reader, ctx, depth, imgChan, rinfo)

	<-token.GetReadTokenChan()
}

func crawlNextPages(reader *goquery.Document, ctx context.Context, depth uint, imgChan chan<- image.ImgDownloadInfo, rinfo model.RobotsInfo) {
	tokenStore := token.GetWriteTokenChan()
	reader.Find("a[href]").Each(func(i int, selection *goquery.Selection) {
		for {
			select {
			case <-ctx.Done():
				return
			case tokenStore <- struct{}{}:
				href := selection.AttrOr("href", "")
				if href != "" {
					CrawlPage(href, depth+1, imgChan, ctx, rinfo)
				}
				time.Sleep(3 * time.Second)
			}
		}
	})
}

func sendImageData(url string, ctx context.Context, reader *goquery.Document, imgChan chan<- image.ImgDownloadInfo) {
	reader.Find("img[src]").Each(func(i int, selection *goquery.Selection) {
		src := selection.AttrOr("src", "")
		altText := selection.AttrOr("altText", "N/A")
		if src != "" {
			select {
			case <-ctx.Done():
				return
			case imgChan <- image.ImgDownloadInfo{Url: src, AltText: altText, RequestUrl: url}:
				log.Printf("Sending image %s\n", src)
			}
		}
	})
}

func canCrawl(url string, depth uint) bool {
	if pageStore == nil {
		log.Fatal("URL HTML page store is not initialized")
	}
	if depth > *model.ParsedFlags.DepthLevel {
		return false
	}

	if pageStore.Contains(url) {
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
	crawlUrlDomain := fmt.Sprintf("%s://%s", parse.Host, parse.Host)
	parse, err = urlpkg.Parse(rinfo.URL)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR parsing url %s: %+v", rinfo.URL, err))
		return false
	}
	robotUrlDomain := fmt.Sprintf("%s://%s", parse.Host, parse.Host)
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
