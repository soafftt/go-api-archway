package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ToStruct[T any](httpResponse *http.Response) (T, error) {
	var result T
	if err := json.NewDecoder(httpResponse.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("decode http response body: %w", err)
	}

	return result, nil
}
