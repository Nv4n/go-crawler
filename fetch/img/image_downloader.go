package img

import (
	"context"
	"fmt"
	"github.com/nv4n/go-crawler/fetch/db"
	"github.com/nv4n/go-crawler/fetch/token"
	"github.com/nv4n/go-crawler/model"
	modelImg "github.com/nv4n/go-crawler/model/image"
	"github.com/nv4n/go-crawler/utils"
	"golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var imageSrcStore *model.UrlStore

func InitImageStore() {
	imageSrcStore = model.InitUrlStore()
}

func FetchImages(imgInfoChan <-chan modelImg.ImgDownloadInfo, ctx context.Context, tokenStore chan<- struct{}) {
	if imageSrcStore == nil {
		log.Fatal("Image URL store is not initialized")
	}

	atomicId := uint64(1)

	for {
		select {
		case <-ctx.Done():
			return
		case imgInfo, ok := <-imgInfoChan:
			if ok {
				tokenStore <- struct{}{}
				go downloadImage(imgInfo.RequestUrl, imgInfo.Url, imgInfo.AltText, atomicId, token.GetReadTokenChan())
				atomicId++
				time.Sleep(3 * time.Second)
			}
		}
	}
}

func downloadImage(requestUrl string, url string, altText string, id uint64, tokenStore <-chan struct{}) {
	if imageSrcStore.Contains(url) {
		return
	}
	match, err := regexp.Match("(http|https)", []byte(url))
	if err != nil || !match {
		url = fmt.Sprintf("%s/%s", requestUrl, url)
	}

	resp, err := http.Get(url)
	imageSrcStore.Add(url)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR downloading %s: %+v", url, err))
		return
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		utils.Warn(fmt.Sprintf("No suitable content-type for %s", url))
		return
	}
	imageSrcStore.Add(url)

	fileFormat := getFileFormat(contentType)
	if fileFormat == "" {
		utils.Warn(fmt.Sprintf("ERROR no suitable img format for: %s", url))
		return
	}

	imgTime := time.Now().Format("2006-01-02_15-01-05")
	filename := fmt.Sprintf("IMG_%s_%d.%s", imgTime, id, fileFormat)

	file, err := os.Create(fmt.Sprintf("uploads/%s", filename))
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR creating file for %s: %+v", url, err))
		return
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR copying file for %s: %+v", url, err))
		return
	}

	metadata := modelImg.DbMetadata{
		Filename:   filename,
		Title:      getTitle(url),
		AltText:    altText,
		Resolution: getResolution(file, fileFormat),
		Format:     fileFormat,
	}
	<-tokenStore
	tokenStoreSend := token.GetWriteTokenChan()
	tokenStoreSend <- struct{}{}
	go db.SaveImage(metadata, tokenStore)
}

func getFileFormat(contentType string) string {
	if strings.Contains(contentType, "svg") {
		return "svg"
	}
	if strings.Contains(contentType, "jpeg") {
		return "jpg"
	}
	if strings.Contains(contentType, "png") {
		return "png"
	}
	if strings.Contains(contentType, "gif") {
		return "gif"
	}
	if strings.Contains(contentType, "webp") {
		return "webp"
	}
	return ""
}

func getResolution(file *os.File, fileFormat string) string {
	invalidValue := "N/A"
	open, err := os.Open(fmt.Sprintf("%s", file.Name()))
	if err != nil {
		utils.Warn(fmt.Sprintf("NAME: %s : %+v", file.Name(), err))
		return invalidValue
	}
	if fileFormat == "jpg" || fileFormat == "png" || fileFormat == "gif" {
		decode, _, err := image.DecodeConfig(open)
		if err != nil {
			utils.Warn(fmt.Sprintf("ERROR decoding config: %+v", err))
			return invalidValue
		}
		return fmt.Sprintf("%d x %d", decode.Width, decode.Height)
	}
	if fileFormat == "webp" {
		decode, err := webp.DecodeConfig(open)
		if err != nil {
			utils.Warn(fmt.Sprintf("ERROR decoding config: %+v", err))
			return invalidValue
		}
		return fmt.Sprintf("%d x %d", decode.Width, decode.Height)
	}
	return invalidValue
}

func getTitle(url string) string {
	return url[strings.LastIndex(url, "/")+1:]
}
