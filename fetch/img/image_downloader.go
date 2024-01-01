package img

import (
	"context"
	"fmt"
	"github.com/nv4n/go-crawler/fetch/store"
	"github.com/nv4n/go-crawler/model"
	"github.com/nv4n/go-crawler/utils"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ImageData struct {
	Filename   string `db:"filename" json:"filename"`
	Title      string `db:"title" json:"title"`
	AltText    string `db:"alt_text" json:"alt_text"`
	Resolution string `db:"resolution" json:"resolution"`
	Format     string `db:"format" json:"format"`
}

var ImageSrcStore *store.UrlStore

func InitImageStore() {
	ImageSrcStore = store.InitUrlStore()
}

func FetchImages(imgInfoChan <-chan model.ImgDownloadInfo, ctx context.Context) {
	atomicId := uint64(1)
	for imgInfo := range imgInfoChan {
		select {
		case <-ctx.Done():
			return
		default:
			go downloadImage(imgInfo.Url, imgInfo.AltText, atomicId)
			atomicId++
		}
	}
}

func downloadImage(url string, altText string, id uint64) {
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
	title := getTitle(url)
	//TODO
	//	Add sql insert
	//	Add image decoder for metadata
	imgId := time.Now().Format("2006-01-02_15-01-05")
	file, err := os.Create(fmt.Sprintf("./uploads/IMG_%s", randId))
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
}

func getFileFormat(contentType string) string {
	if strings.Contains(contentType, "svg") {
		return "svg"
	}
	if strings.Contains(contentType, "jpeg") {
		return "jpeg"
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

func getTitle(url string) string {
	return url[strings.LastIndex(url, "/")+1:]
}
