package vbmmodels

import "time"

type VbmLicense struct {
	LicenseID             string    `json:"licenseID"`
	Email                 string    `json:"email"`
	Status                string    `json:"status"`
	ExpirationDate        string    `json:"expirationDate"`
	Type                  string    `json:"type"`
	LicensedTo            string    `json:"licensedTo"`
	ContactPerson         string    `json:"contactPerson"`
	TotalNumber           int       `json:"totalNumber"`
	UsedNumber            int       `json:"usedNumber"`
	NewNumber             int       `json:"newNumber"`
	SupportID             string    `json:"supportID"`
	SupportExpirationDate time.Time `json:"supportExpirationDate"`
}
