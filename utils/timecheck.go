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

	if ntime.After(otime) && len(ctime) > 0 {
		return true
	} else {
		return false
	}
}
