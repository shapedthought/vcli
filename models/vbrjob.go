package models

import "time"

type VbrJob struct {
	Data []struct {
		ID             string    `json:"id"`
		Name           string    `json:"name"`
		Type           string    `json:"type"`
		Description    string    `json:"description"`
		Status         string    `json:"status"`
		LastRun        time.Time `json:"lastRun"`
		LastResult     string    `json:"lastResult"`
		NextRun        time.Time `json:"nextRun"`
		Workload       string    `json:"workload"`
		RepositoryID   string    `json:"repositoryId"`
		RepositoryName string    `json:"repositoryName"`
		ObjectsCount   int       `json:"objectsCount"`
		SessionID      string    `json:"sessionId"`
	} `json:"data"`
	Pagination struct {
		Total int `json:"total"`
		Count int `json:"count"`
		Skip  int `json:"skip"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}
