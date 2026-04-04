package response

import (
	"encoding/json"
	commonModel "gateway/common/model"
	"net/http"
)

func HandErrorResponse(w http.ResponseWriter, status int, message string, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(commonModel.ErrorResponse{Message: message, Detail: detail})
}
