package vbrmodels

type Proxies struct {
	Data []struct {
		Server struct {
			HostID                string `json:"hostId"`
			TransportMode         string `json:"transportMode"`
			FailoverToNetwork     bool   `json:"failoverToNetwork"`
			HostToProxyEncryption bool   `json:"hostToProxyEncryption"`
			ConnectedDatastores   struct {
				AutoSelect bool `json:"autoSelect"`
				Datastores []struct {
					Datastore struct {
						HostName string `json:"hostName"`
						Name     string `json:"name"`
						Type     string `json:"type"`
						ObjectID string `json:"objectId"`
					} `json:"datastore"`
					VMCount int `json:"vmCount"`
				} `json:"datastores"`
			} `json:"connectedDatastores"`
			MaxTaskCount int `json:"maxTaskCount"`
		} `json:"server"`
		Type        string `json:"type"`
		Name        string `json:"name"`
		Id          string `json:"id"`
		Description string `json:"description"`
	} `json:"data"`
	Pagination struct {
		Total int `json:"total"`
		Count int `json:"count"`
		Skip  int `json:"skip"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}
