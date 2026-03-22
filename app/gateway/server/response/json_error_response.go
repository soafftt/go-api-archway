package response

import (
	"encoding/json"
	"gateway/model"
	"net/http"

	"github.com/google/wire"
)

type WriteError func(w http.ResponseWriter, status int, message string, detail string)

type JsonErrorResponse struct {
	WriteResponse WriteError
}

func NewJsonErrorWriter() *JsonErrorResponse {
	return &JsonErrorResponse{
		WriteResponse: func(w http.ResponseWriter, status int, message, detail string) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)

			json.NewEncoder(w).Encode(model.NewErrorResponse(message, detail))
		},
	}
}

var JsonErrorResponseSet = wire.NewSet(
	NewJsonErrorWriter,
)
