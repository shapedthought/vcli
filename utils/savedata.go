package utils

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v2"
)

func SaveData[T any](data *T, name string) {
	file, err := yaml.Marshal(&data)
	IsErr(err)

	_ = os.WriteFile(name+".yaml", file, 0644)
}

func SaveJson[T any](data *T, name string) {
	file, err := json.MarshalIndent(&data, "", "    ")
	IsErr(err)

	_ = os.WriteFile(name+".json", file, 0644)
}
