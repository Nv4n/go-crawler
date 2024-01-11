package model

type CliFlags struct {
	Url           *string `validate:"required,url"`
	ExternalLinks *bool
	Spa           *bool
	Timeout       *int  `validate:"min=1,max=10"`
	Goroutines    *uint `validate:"min=15,max=100"`
	DepthLevel    *uint `validate:"min=1,max=10"`
}

var ParsedFlags CliFlags
