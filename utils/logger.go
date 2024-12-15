package utils

import (
	"log"
	"os"
)

func Warn(msg string) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetPrefix("Warn: ")
	log.Println(msg)

	log.SetFlags(log.LstdFlags)
	log.SetPrefix("")
	log.SetOutput(os.Stderr)
}
