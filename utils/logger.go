package utils

import "log"

func Warn(msg string) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetPrefix("Warn: ")
	log.Println(msg)
}
