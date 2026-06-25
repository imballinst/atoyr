package utils

import (
	"atoyr/server/internal/models"
	"encoding/json"
	"time"
)

func ConvertDbJsonToNestedStringArray(j models.JSON) ([][]string, error) {
	var result [][]string
	err := json.Unmarshal([]byte(j), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ConvertNestedStringArrayToDbJson(arr [][]string) (models.JSON, error) {
	data, err := json.Marshal(arr)
	if err != nil {
		return nil, err
	}
	return models.JSON(data), nil
}

// Number is any numeric type this function accepts.
type Number interface {
	int32 | float32
}

// ToDuration converts a value to a time.Duration.
// - For float32 < 1, treats the value as milliseconds.
// - Otherwise, treats the value as seconds.
func ToDuration[T Number](v T) time.Duration {
	f := float64(v)

	if f < 1 {
		return time.Duration(f * float64(time.Millisecond))
	}
	return time.Duration(f * float64(time.Second))
}
