package test

import (
	"fmt"
	dbpkg "github.com/nv4n/go-crawler/fetch/db"
	"github.com/nv4n/go-crawler/model/image"
	"testing"
	"time"
)

func TestDbConnection(t *testing.T) {
	dbpkg.InitDb()
	defer dbpkg.CloseDb()
	err := dbpkg.Ping()
	if err != nil {
		t.Errorf("Db can't be pinged: %+v", err)
	}
}

func ExampleImgUpload() {
	dbpkg.InitDb()
	defer dbpkg.CloseDb()
	tokenStore := make(chan struct{}, 5)
	tokenStore <- struct{}{}
	exampleInfo := image.DbMetadata{Filename: "Filename", Title: "Title", AltText: "N/A", Resolution: "90x180", Format: "JPEG"}
	dbpkg.SaveImage(exampleInfo, tokenStore)
	time.Sleep(2 * time.Second)
	imgs := dbpkg.GetAllImages()
	fmt.Println(imgs[len(imgs)-1])
	affected, err := dbpkg.DeleteLastRow()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(affected)
	// Output:
	// {Filename Title N/A 90x180 JPEG}
	// 1
}
