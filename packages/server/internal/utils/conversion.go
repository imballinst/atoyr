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

func ConvertStringArrayToTimestampJSON(arr [][]string) (models.JSON, error) {
	data, err := json.Marshal(arr)
	if err != nil {
		return nil, err
	}
	return models.JSON(data), nil
}
