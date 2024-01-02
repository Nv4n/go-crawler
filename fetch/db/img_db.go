package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nv4n/go-crawler/model"
	"github.com/nv4n/go-crawler/utils"
	"log"
	"os"
)

var db *sql.DB

func InitDb() {
	pass := os.Getenv("DB_PASS")
	initDb, err := sql.Open("postgres", fmt.Sprintf("postgresql:postgres@%s/localhost:5432/golang_crawler", pass))
	if err != nil {
		log.Fatalf("Can't open initDb: %+v", err)
	}
	db = initDb
}

func dbInitCheck() {
	if db == nil {
		log.Fatal("DB is not initialized")
	}
}
func CloseDb() {
	dbInitCheck()
	db.Close()
}

func SaveImage(info model.ImageData, tokenStore <-chan struct{}) {
	dbInitCheck()
	_, err := db.Exec("INSERT INTO public.image_metadata(filename, title, alt_text, resolution, format) VALUES ($1,$2,$3,$4,$5)", info.Filename, info.Title, info.AltText, info.Resolution, info.Format)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR create new log in db: %+v", err))
	}
	<-tokenStore
}

func GetAllImages() []model.ImageData {
	dbInitCheck()
	rows, err := db.Query("SELECT * FROM public.image_metadata")
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR fetching all image metadata: %+v", err))
		return nil
	}
	defer rows.Close()
	var imageMetadata []model.ImageData
	for rows.Next() {
		metadata := model.ImageData{}
		err = rows.Scan(&metadata.Filename, &metadata.Title, &metadata.AltText, &metadata.Resolution, &metadata.Format)
		if err != nil {
			utils.Warn(fmt.Sprintf("ERROR scanning image: %+v", err))
			return imageMetadata
		}
		imageMetadata = append(imageMetadata, metadata)
	}
	return imageMetadata
}
