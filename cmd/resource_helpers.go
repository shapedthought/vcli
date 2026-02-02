package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/shapedthought/vcli/models"
	"github.com/shapedthought/vcli/vhttp"
)

// findResourceInList fetches a list endpoint and finds a resource by matching
// a field value. Returns the matching raw JSON entry and its ID.
// Works for any resource type: jobs match on "name", encryption passwords on "hint", etc.
func findResourceInList(listEndpoint, matchField, matchValue string, profile models.Profile) (json.RawMessage, string, error) {
	type ListResponse struct {
		Data []json.RawMessage `json:"data"`
	}

	list := vhttp.GetData[ListResponse](listEndpoint, profile)

	for _, raw := range list.Data {
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		if fmt.Sprintf("%v", m[matchField]) == matchValue {
			id, _ := m["id"].(string)
			return raw, id, nil
		}
	}

	return nil, "", fmt.Errorf("resource with %s '%s' not found at %s", matchField, matchValue, listEndpoint)
}
