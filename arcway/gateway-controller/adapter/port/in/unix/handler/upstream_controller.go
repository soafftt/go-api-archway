package handler

import (
	in2 "core/adapter/in"
	"encoding/json"
	"gateway/controller/adapter/port/in/unix"
	"gateway/controller/application/port/in"
	"gateway/controller/application/port/in/dto"
	"log"
	"net/http"
)

// hander interface
type UpstreamRouter unix.Router

type upStreamHandler struct {
	upStreamLookupUseCase in.UpstreamLookupUseCase
}

func NewUpStreamHandler(upStreamUseCase in.UpstreamLookupUseCase) UpstreamRouter {
	return &upStreamHandler{
		upStreamLookupUseCase: upStreamUseCase,
	}
}

func (h *upStreamHandler) Routes() []unix.Route {
	return []unix.Route{
		{Method: "GET", Path: "/v1/upstream", Handler: h.upStream},
	}
}

func (h *upStreamHandler) upStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	request, err := dto.NewUpStreamLookupRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(in2.NewErrorResponse(err)); err != nil {
			log.Printf("failed to encode upstream lookup request error: %v", err)
		}
		return
	}

	result := h.upStreamLookupUseCase.LookUp(request)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(in2.NewErrorResponse(result.Error)); err != nil {
			log.Printf("failed to encode upstream lookup error response: %v", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(result.Info); err != nil {
		log.Printf("failed to encode upstream lookup result: %v", err)
	}
}
