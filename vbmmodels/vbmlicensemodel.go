package vbmmodels

import "time"

type VbmLicense struct {
	LicenseID             string    `json:"licenseID" yaml:"licenseID"`
	Email                 string    `json:"email" yaml:"email"`
	Status                string    `json:"status" yaml:"status"`
	ExpirationDate        string    `json:"expirationDate" yaml:"expirationDate"`
	Type                  string    `json:"type" yaml:"type"`
	LicensedTo            string    `json:"licensedTo" yaml:"licensedTo"`
	ContactPerson         string    `json:"contactPerson" yaml:"contactPerson"`
	TotalNumber           int       `json:"totalNumber" yaml:"totalNumber"`
	UsedNumber            int       `json:"usedNumber" yaml:"userNumber"`
	NewNumber             int       `json:"newNumber" yaml:"newNumber"`
	SupportID             string    `json:"supportID" yaml:"supportID"`
	SupportExpirationDate time.Time `json:"supportExpirationDate" yaml:"supportExpirationDate"`
}
