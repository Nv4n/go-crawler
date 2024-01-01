package img

import (
	"context"
	"fmt"
	"github.com/nv4n/go-crawler/fetch/url"
	"github.com/nv4n/go-crawler/model"
	"github.com/nv4n/go-crawler/utils"
	"golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var ImageSrcStore *url.Store

func InitImageStore() {
	ImageSrcStore = url.InitUrlStore()
}

func FetchImages(imgInfoChan <-chan model.ImgDownloadInfo, ctx context.Context, tokenStore chan struct{}) {
	atomicId := uint64(1)

	for imgInfo := range imgInfoChan {
		select {
		case <-ctx.Done():
			return
		default:
			tokenStore <- struct{}{}
			go downloadImage(imgInfo.Url, imgInfo.AltText, atomicId, tokenStore)
			atomicId++
		}
	}
}

func downloadImage(url string, altText string, id uint64, tokenStore <-chan struct{}) {
	resp, err := http.Get(url)
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

	fileFormat := getFileFormat(contentType)

	imgTime := time.Now().Format("2006-01-02_15-01-05")
	filename := fmt.Sprintf("IMG_%s_%d.%s", imgTime, id, fileFormat)

	file, err := os.Create(fmt.Sprintf("../../uploads/%s", filename))
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

	metadata := model.ImageData{
		Filename:   filename,
		Title:      getTitle(url),
		AltText:    altText,
		Resolution: getResolution(file, fileFormat),
		Format:     fileFormat,
	}

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
	if fileFormat == "jpg" || fileFormat == "png" || fileFormat == "gif" {
		decode, _, err := image.DecodeConfig(file)
		if err != nil {
			utils.Warn(fmt.Sprintf("ERROR decoding config: %+v", err))
			return invalidValue
		}
		return fmt.Sprintf("%d x %d", decode.Width, decode.Height)
	}
	if fileFormat == "webp" {
		decode, err := webp.DecodeConfig(file)
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
