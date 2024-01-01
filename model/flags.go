package model

type CliFlags struct {
	Url           *string `validate:"required,url"`
	ExternalLinks *bool
	Spa           *bool
	Timeout       *int  `validate:"min=1,max=10"`
	Goroutines    *uint `validate:"min=10,max=60"`
	DepthLevel    *uint `validate:"min=1,max=10"`
}
