package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nv4n/go-crawler/model/image"
	"github.com/nv4n/go-crawler/utils"
	"log"
	"os"
)

var db *sql.DB

func InitDb() {
	pass := os.Getenv("DB_PASS")
	initDb, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres:%s@localhost:5432/golang_crawler?sslmode=disable", pass))
	if err != nil {
		log.Fatalf("Can't open initDb: %+v", err)
	}
	db = initDb
}
func Ping() error {
	return db.Ping()
}
func DeleteLastRow() (int64, error) {
	r, err := db.Exec("DELETE FROM image_metadata WHERE id IN (SELECT id FROM image_metadata ORDER BY id DESC FETCH FIRST ROW ONLY )")
	if err != nil {
		return -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, err
	}
	return affected, nil
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

func SaveImage(info image.DbMetadata, tokenStore <-chan struct{}) {
	dbInitCheck()
	_, err := db.Exec("INSERT INTO public.image_metadata(filename, title, alt_text, resolution, format) VALUES ($1,$2,$3,$4,$5)", info.Filename, info.Title, info.AltText, info.Resolution, info.Format)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR create new log in db: %+v", err))
	}
	<-tokenStore
}

func GetAllImages() []image.DbMetadata {
	dbInitCheck()
	rows, err := db.Query("SELECT filename,title,alt_text,resolution,format FROM public.image_metadata")
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR fetching all image metadata: %+v", err))
		return nil
	}
	defer rows.Close()
	var imageMetadata []image.DbMetadata
	for rows.Next() {
		metadata := image.DbMetadata{}
		err = rows.Scan(&metadata.Filename, &metadata.Title, &metadata.AltText, &metadata.Resolution, &metadata.Format)
		if err != nil {
			utils.Warn(fmt.Sprintf("ERROR scanning image: %+v", err))
			return imageMetadata
		}
		imageMetadata = append(imageMetadata, metadata)
	}
	return imageMetadata
}
