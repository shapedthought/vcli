package vbmmodels

type VbmProxies []VbmProxy

type VbmProxy struct {
	Type                    string `json:"type"`
	UseInternetProxy        bool   `json:"useInternetProxy"`
	InternetProxyType       string `json:"internetProxyType"`
	ID                      string `json:"id"`
	HostName                string `json:"hostName"`
	Description             string `json:"description"`
	Port                    int    `json:"port"`
	ThreadsNumber           int    `json:"threadsNumber"`
	EnableNetworkThrottling bool   `json:"enableNetworkThrottling"`
	Status                  string `json:"status"`
	Links                   struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Repositories struct {
			Href string `json:"href"`
		} `json:"repositories"`
	} `json:"_links"`
}
