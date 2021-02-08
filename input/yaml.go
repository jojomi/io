package input

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
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
