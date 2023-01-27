package input

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func GetYamlFromFile(filename string) (map[string]interface{}, error) {
	var data map[string]interface{}

	// read dynamic yaml
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(source, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetJSONFromFile(filename string) (map[string]interface{}, error) {
	var data map[string]interface{}

	// read dynamic yaml
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(source, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
