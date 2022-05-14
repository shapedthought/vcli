package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

func CheckTime(ctime string) bool {
	otime, err := time.Parse(time.RFC3339, ctime)
	IsErr(err)

	ntime := time.Now()

	if ntime.After(otime) {
		return true
	} else {
		return false
	}
}

func UpdateTime() {
	headers := ReadHeader()

	ntime := time.Now().Add(time.Minute * 15).Format(time.RFC3339)

	headers.Expires = ntime

	d, err := json.Marshal(headers)
	IsErr(err)

	if err := ioutil.WriteFile("headers.json", d, 0655); err != nil {
		log.Fatalf("Could not save headers file - %v", err)
	}
}
