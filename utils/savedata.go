package utils

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func SaveData[T any](data *T, name string) {
	file, err := yaml.Marshal(&data)
	IsErr(err)

	_ = ioutil.WriteFile(name+".yaml", file, 0644)
}

func SaveJson[T any](data *T, name string) {
	file, err := json.MarshalIndent(&data, "", "    ")
	IsErr(err)

	_ = ioutil.WriteFile(name+".json", file, 0644)
}
