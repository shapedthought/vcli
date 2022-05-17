package vbmmodels

type VbmJobs []struct {
	Description    string `json:"description" yaml:"description"`
	BackupType     string `json:"backupType" yaml:"backupType"`
	SchedulePolicy struct {
		ScheduleEnabled     bool   `json:"scheduleEnabled" yaml:"scheduleEnabled"`
		BackupWindowEnabled bool   `json:"backupWindowEnabled" yaml:"backupWindowEnabled"`
		Type                string `json:"type" yaml:"type"`
		DailyType           string `json:"dailyType" yaml:"dailyType"`
		DailyTime           string `json:"dailyTime" yaml:"dailyTime"`
		RetryEnabled        bool   `json:"retryEnabled" yaml:"retryEnabled"`
		RetryNumber         int    `json:"retryNumber" yaml:"retryNumber"`
		RetryWaitInterval   int    `json:"retryWaitInterval" yaml:"retryWaitInterval"`
	} `json:"schedulePolicy" yaml:"schedulePolicy"`
	ID           string `json:"id" yaml:"id"`
	RepositoryID string `json:"repositoryId" yaml:"repositoryId"`
	Name         string `json:"name" yaml:"name"`
	LastRun      string `json:"lastRun,omitempty" yaml:"lastRun"`
	NextRun      string `json:"nextRun" yaml:"nextRun"`
	IsEnabled    bool   `json:"isEnabled" yaml:"isEnabled"`
	LastStatus   string `json:"lastStatus" yaml:"lastStatus"`
	Links        struct {
		Self struct {
			Href string `json:"href" yaml:"href"`
		} `json:"self" yaml:"self"`
		CopyJob struct {
			Href string `json:"href" yaml:"href"`
		} `json:"copyJob" yaml:"copyJob"`
		Organization struct {
			Href string `json:"href" yaml:"href"`
		} `json:"organization" yaml:"organization"`
		BackupRepository struct {
			Href string `json:"href" yaml:"href"`
		} `json:"backupRepository" yaml:"backupRepository"`
		Jobsessions struct {
			Href string `json:"href" yaml:"href"`
		} `json:"jobsessions" yaml:"jobsessions"`
		ExcludedItems struct {
			Href string `json:"href" yaml:"href"`
		} `json:"excludedItems" yaml:"excludedItems"`
		SelectedItems struct {
			Href string `json:"href" yaml:"href"`
		} `json:"selectedItems" yaml:"selectedItems"`
	} `json:"_links" yaml:"_links"`
}
