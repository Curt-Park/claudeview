package config

import (
	"encoding/json"
	"os"
)

// loadJSON reads a JSON file at path and unmarshals it into T.
// Returns the zero value of T (and nil error) if the file does not exist.
func loadJSON[T any](path string) (T, error) {
	var zero T
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return zero, nil
		}
		return zero, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return zero, err
	}
	return v, nil
}
