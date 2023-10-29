package http

import (
	"encoding/json"
	"log"
	"net/http"
)

type jsonResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func RespondSuccess(w http.ResponseWriter, data interface{}) {
	response := &jsonResponse{
		Success: true,
		Code:    http.StatusOK,
		Message: "Success",
		Data:    data,
	}
	sendJSONResponse(w, http.StatusOK, response)
}

// RespondError sends an error JSON response.
func RespondError(w http.ResponseWriter, code int, message string, err error) {
	if err != nil {
		log.Printf("Error: %v", err)
	}
	response := &jsonResponse{
		Success: false,
		Code:    code,
		Message: message,
	}
	sendJSONResponse(w, code, response)
}

func sendJSONResponse(w http.ResponseWriter, code int, response *jsonResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
