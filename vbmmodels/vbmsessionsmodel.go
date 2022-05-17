package vbmmodels

import "time"

type VbmSessions struct {
	Links struct {
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Limit   int `json:"limit"`
	Offset  int `json:"offset"`
	Results []struct {
		Links struct {
			Job struct {
				Href string `json:"href"`
			} `json:"job"`
			Log struct {
				Href string `json:"href"`
			} `json:"log"`
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"_links"`
		CreationTime time.Time `json:"creationTime"`
		Details      string    `json:"details"`
		EndTime      time.Time `json:"endTime"`
		ID           string    `json:"id"`
		Progress     int       `json:"progress"`
		RetryCount   int       `json:"retryCount"`
		Statistics   struct {
			Bottleneck            string `json:"bottleneck"`
			ProcessedObjects      int    `json:"processedObjects"`
			ProcessingRateBytesPS int    `json:"processingRateBytesPS"`
			ProcessingRateItemsPS int    `json:"processingRateItemsPS"`
			ReadRateBytesPS       int    `json:"readRateBytesPS"`
			TransferredDataBytes  int    `json:"transferredDataBytes"`
			WriteRateBytesPS      int    `json:"writeRateBytesPS"`
		} `json:"statistics"`
		Status string `json:"status"`
	} `json:"results"`
	SetID string `json:"setId"`
}
