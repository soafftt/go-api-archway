package response

import (
	"encoding/json"
	"net/http"
)

func HandErrorResponse(w http.ResponseWriter, status int, message string, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(model.ErrorResponse{Message: message, Detail: detail})
}
