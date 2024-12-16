package image

type DbMetadata struct {
	Id         int    `db:"id" json:"id"`
	Filename   string `db:"filename" json:"filename"`
	Title      string `db:"title" json:"title"`
	AltText    string `db:"alt_text" json:"alt_text"`
	Resolution string `db:"resolution" json:"resolution"`
	Format     string `db:"format" json:"format"`
}

type DbFilter struct {
	Title   string `db:"title" json:"title"`
	AltText string `db:"alt_text" json:"alt_text"`
	Format  string `db:"format" json:"format"`
}
