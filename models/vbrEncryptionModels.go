package models

// VbrEncryptionPasswordGet represents an encryption password from the VBR API.
// Only metadata is stored — never actual password values.
type VbrEncryptionPasswordGet struct {
	ID               string `json:"id" yaml:"id"`
	Hint             string `json:"hint" yaml:"hint"`
	ModificationTime string `json:"modificationTime" yaml:"modificationTime"`
	UniqueID         string `json:"uniqueId" yaml:"uniqueId"`
	IsImported       bool   `json:"isImported" yaml:"isImported"`
}

// VbrEncryptionPasswordList represents the list response from the encryption passwords endpoint
type VbrEncryptionPasswordList struct {
	Data       []VbrEncryptionPasswordGet `json:"data"`
	Pagination map[string]interface{}     `json:"pagination,omitempty"`
}

// VbrKmsServerGet represents a KMS server from the VBR API
type VbrKmsServerGet struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Type        string `json:"type" yaml:"type"`
}

// VbrKmsServerList represents the list response from the KMS servers endpoint
type VbrKmsServerList struct {
	Data       []VbrKmsServerGet      `json:"data"`
	Pagination map[string]interface{} `json:"pagination,omitempty"`
}
