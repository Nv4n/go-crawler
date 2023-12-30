package img

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nv4n/go-crawler/fetch/store"
	"github.com/nv4n/go-crawler/utils"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
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
	ImageSrcStore := store.InitUrlStore()
}

func FetchImages(imgUrls []string) {
	for _, url := range imgUrls {
		if ImageSrcStore.Contains(url) {
			continue
		}

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

func DownloadImage(url string, altText string) {
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
	randId := uuid.New().String()
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
	stat, err := file.Stat()

	if err != nil {
		return
	}
}
