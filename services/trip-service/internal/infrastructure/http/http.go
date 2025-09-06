// Package http
package http

import (
	"encoding/json"
	"log"
	"net/http"

	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/types"
)

type HTTPHandler struct {
	Service domain.TripService
}

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func (h *HTTPHandler) HandleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ctx := r.Context()
	t, err := h.Service.GetRoute(ctx, &reqBody.Pickup, &reqBody.Destination)
	if err != nil {
		log.Println(err)
		return
	}

	writeJSON(w, http.StatusOK, t)
}
