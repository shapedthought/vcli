package utils

import (
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
