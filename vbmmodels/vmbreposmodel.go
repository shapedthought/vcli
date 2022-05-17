package vbmmodels

type VbmRepos []VbmRepo

type VbmRepo struct {
	IsOutOfSync            bool   `json:"isOutOfSync"`
	CapacityBytes          int64  `json:"capacityBytes"`
	FreeSpaceBytes         int64  `json:"freeSpaceBytes"`
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Description            string `json:"description"`
	Path                   string `json:"path,omitempty"`
	RetentionType          string `json:"retentionType"`
	RetentionPeriodType    string `json:"retentionPeriodType"`
	YearlyRetentionPeriod  string `json:"yearlyRetentionPeriod,omitempty"`
	RetentionFrequencyType string `json:"retentionFrequencyType"`
	DailyTime              string `json:"dailyTime"`
	DailyType              string `json:"dailyType"`
	ProxyID                string `json:"proxyId"`
	Links                  struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Proxy struct {
			Href string `json:"href"`
		} `json:"proxy"`
	} `json:"_links"`
	ObjectStorageID                string `json:"objectStorageId,omitempty"`
	ObjectStorageCachePath         string `json:"objectStorageCachePath,omitempty"`
	ObjectStorageEncryptionEnabled bool   `json:"objectStorageEncryptionEnabled,omitempty"`
	DailyRetentionPeriod           int    `json:"dailyRetentionPeriod,omitempty"`
	IsLongTerm                     bool   `json:"isLongTerm,omitempty"`
}
