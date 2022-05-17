package vbmmodels

import "time"

type VbmOrg struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	IsBackedup      bool      `json:"isBackedup"`
	FirstBackuptime time.Time `json:"firstBackuptime"`
	LastBackuptime  time.Time `json:"lastBackuptime"`
	Actions         struct {
		Explore struct {
			URI    string `json:"uri"`
			Method string `json:"method"`
		} `json:"explore"`
	} `json:"_actions"`
}

type VbmOrgs []struct {
	IsTeamsOnline          bool `json:"isTeamsOnline"`
	ExchangeOnlineSettings struct {
		UseApplicationOnlyAuth           bool   `json:"useApplicationOnlyAuth"`
		OfficeOrganizationName           string `json:"officeOrganizationName"`
		Account                          string `json:"account"`
		GrantAdminAccess                 bool   `json:"grantAdminAccess"`
		UseMfa                           bool   `json:"useMfa"`
		ApplicationID                    string `json:"applicationId"`
		ApplicationCertificateThumbprint string `json:"applicationCertificateThumbprint"`
	} `json:"exchangeOnlineSettings"`
	SharePointOnlineSettings struct {
		UseApplicationOnlyAuth           bool   `json:"useApplicationOnlyAuth"`
		OfficeOrganizationName           string `json:"officeOrganizationName"`
		SharePointSaveAllWebParts        bool   `json:"sharePointSaveAllWebParts"`
		Account                          string `json:"account"`
		GrantAdminAccess                 bool   `json:"grantAdminAccess"`
		UseMfa                           bool   `json:"useMfa"`
		ApplicationID                    string `json:"applicationId"`
		ApplicationCertificateThumbprint string `json:"applicationCertificateThumbprint"`
	} `json:"sharePointOnlineSettings"`
	IsExchangeOnline   bool      `json:"isExchangeOnline"`
	IsSharePointOnline bool      `json:"isSharePointOnline"`
	Type               string    `json:"type"`
	Region             string    `json:"region"`
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	OfficeName         string    `json:"officeName"`
	IsBackedup         bool      `json:"isBackedup"`
	FirstBackuptime    time.Time `json:"firstBackuptime"`
	LastBackuptime     time.Time `json:"lastBackuptime"`
	Links              struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Jobs struct {
			Href string `json:"href"`
		} `json:"jobs"`
		Groups struct {
			Href string `json:"href"`
		} `json:"groups"`
		Users struct {
			Href string `json:"href"`
		} `json:"users"`
		Sites struct {
			Href string `json:"href"`
		} `json:"sites"`
		Teams struct {
			Href string `json:"href"`
		} `json:"teams"`
		UsedRepositories struct {
			Href string `json:"href"`
		} `json:"usedRepositories"`
	} `json:"_links"`
	Actions struct {
	} `json:"_actions"`
}
