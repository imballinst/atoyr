package utils

import (
	"atoyr/server/internal/models"
	"encoding/json"
)

func ConvertTimestampJSONToStringArray(j models.JSON) ([][]string, error) {
	var result [][]string
	err := json.Unmarshal([]byte(j), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
