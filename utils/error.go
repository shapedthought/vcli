package utils

import "log"

func IsErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
