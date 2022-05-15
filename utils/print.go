package utils

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

func PrintYaml[T any](data *T) {
	file, err := yaml.Marshal(&data)
	IsErr(err)

	fmt.Println(string(file))
}

func PrintJson[T any](data *T) {
	file, err := json.MarshalIndent(&data, "", "    ")
	IsErr(err)

	fmt.Println(string(file))
}
