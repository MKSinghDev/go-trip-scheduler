package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"ride-sharing/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	// Validation
	if reqBody.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		log.Println(err)
	}

	reader := bytes.NewReader(data)

	// TODO: Call trip service
	res, err := http.Post("http://trip-service:8083/preview", "application/json", reader)
	if err != nil {
		log.Printf("Error on get call: %v", err)
		response := contracts.APIResponse{Error: &contracts.APIError{Code: "Internal Server Error", Message: fmt.Sprint(err)}}
		writeJSON(w, http.StatusInternalServerError, response)
		return
	}
	defer res.Body.Close()

	var respBody any
	if err := json.NewDecoder(res.Body).Decode(&respBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	response := contracts.APIResponse{Data: respBody}
	writeJSON(w, http.StatusCreated, response)
}
