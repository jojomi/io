package input

import (
	"encoding/csv"
	"io"
	"os"
)

func GetCSVFromFile(filename string) ([]map[string]interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return csvToMap(f)
}

func csvToMap(reader io.Reader) ([]map[string]interface{}, error) {
	r := csv.NewReader(reader)
	rows := []map[string]interface{}{}
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return rows, err
		}
		if len(header) == 0 {
			header = record
			continue
		}

		dict := map[string]interface{}{}
		for i := range header {
			dict[header[i]] = record[i]
		}
		rows = append(rows, dict)
	}
	return rows, nil
}
