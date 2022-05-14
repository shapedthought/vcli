package vbrmodels

import "time"

type Sessions struct {
	Data []struct {
		ID              string    `json:"id" yaml:"id"`
		Name            string    `json:"name" yaml:"name"`
		ActivityID      string    `json:"activityId" yaml:"activityId"`
		SessionType     string    `json:"sessionType" yaml:"sessionType"`
		CreationTime    time.Time `json:"creationTime" yaml:"creationTime"`
		EndTime         time.Time `json:"endTime" yaml:"endTime"`
		State           string    `json:"state" yaml:"state"`
		ProgressPercent int       `json:"progressPercent" yaml:"progressPercent"`
		Result          struct {
			Result     string `json:"result" yaml:"result"`
			Message    string `json:"message" yaml:"message"`
			IsCanceled bool   `json:"isCanceled" yaml:"isCanceled"`
		} `json:"result" yaml:"result"`
		ResourceID        string `json:"resourceId" yaml:"resourceId"`
		ResourceReference string `json:"resourceReference" yaml:"resourceReference"`
		ParentSessionID   string `json:"parentSessionId" yaml:"parentSessionId"`
		Usn               int    `json:"usn" yaml:"usn"`
	} `json:"data" yaml:"data"`
	Pagination struct {
		Total int `json:"total" yaml:"total"`
		Count int `json:"count" yaml:"count"`
		Skip  int `json:"skip" yaml:"skip"`
		Limit int `json:"limit" yaml:"limit"`
	} `json:"pagination" yaml:"pagination"`
}
