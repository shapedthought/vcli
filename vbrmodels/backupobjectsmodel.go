package vbrmodels

type BackupObjects struct {
	Data []struct {
		ObjectID           string `json:"objectId" yaml:"objectId"`
		ViType             string `json:"viType" yaml:"viType"`
		Path               string `json:"path" yaml:"path"`
		ID                 string `json:"id" yaml:"id"`
		Type               string `json:"type" yaml:"type"`
		RestorePointsCount int    `json:"restorePointsCount" yaml:"restorePointsCount"`
		PlatformName       string `json:"platformName" yaml:"platformName"`
		PlatformId         string `json:"platformId" yaml:"platformId"`
		Name               string `json:"name" yaml:"name"`
	} `json:"data" yaml:"data"`
	Pagination struct {
		Total int `json:"total" yaml:"total"`
		Count int `json:"count" yaml:"count"`
		Skip  int `json:"skip" yaml:"skip"`
		Limit int `json:"limit" yaml:"limit"`
	} `json:"pagination" yaml:"pagination"`
}
