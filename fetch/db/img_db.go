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

var Db *sql.DB

func InitDb() {
	pass := os.Getenv("DB_PASS")
	db, err := sql.Open("postgres", fmt.Sprintf("postgresql:postgres@%s/localhost:5432/golang_crawler", pass))
	if err != nil {
		log.Fatalf("Can't open db: %+v", err)
	}
	Db = db
}

func SaveImage(info model.ImageData) {
	_, err := Db.Exec("INSERT INTO public.image_metadata(filename, title, alt_text, resolution, format) VALUES ($1,$2,$3,$4,$5)", info.Filename, info.Title, info.AltText, info.Resolution, info.Format)
	if err != nil {
		utils.Warn(fmt.Sprintf("ERROR create new log in db: %+v", err))
	}
}
