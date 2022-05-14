package vbrmodels

type Hosts struct {
	Data []struct {
		InventoryObject struct {
			HostName string `json:"hostName"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			ObjectID string `json:"objectId"`
		} `json:"inventoryObject"`
		Size string `json:"size"`
	} `json:"data"`
	Pagination struct {
		Total int `json:"total"`
		Count int `json:"count"`
		Skip  int `json:"skip"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}

type Inventory struct {
	Data []struct {
		InventoryObject struct {
			HostName string `json:"hostName"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			ObjectID string `json:"objectId"`
		} `json:"inventoryObject"`
		Size string `json:"size"`
	} `json:"data"`
	Pagination struct {
		Total int `json:"total"`
		Count int `json:"count"`
		Skip  int `json:"skip"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}
