package utils

import (
	"time"
)

func CheckTime(ctime string, api string) bool {
	var otime time.Time
	var err error

	if api == "vbr" {
		otime, err = time.Parse(time.RFC3339, ctime)
	} else if api == "vbm365" {
		otime, err = time.Parse(time.RFC1123, ctime)
	}

	IsErr(err)

	ntime := time.Now()

	if ntime.After(otime) {
		return true
	} else {
		return false
	}
}

// func UpdateTime() {
// 	headers := ReadHeader()

// 	ntime := time.Now().Add(time.Minute * 15).Format(time.RFC3339)

// 	headers.Expires = ntime

// 	d, err := json.Marshal(headers)
// 	IsErr(err)

// 	if err := ioutil.WriteFile("headers.json", d, 0655); err != nil {
// 		log.Fatalf("Could not save headers file - %v", err)
// 	}
// }
