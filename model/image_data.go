package model

type ImageData struct {
	Filename   string `db:"filename" json:"filename"`
	Title      string `db:"title" json:"title"`
	AltText    string `db:"alt_text" json:"alt_text"`
	Resolution string `db:"resolution" json:"resolution"`
	Format     string `db:"format" json:"format"`
}
