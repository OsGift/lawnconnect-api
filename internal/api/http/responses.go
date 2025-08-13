package http

import (
	"encoding/json"
	"log"
	"net/http"
)

// GeneralResponse defines a standard JSON response format.
type GeneralResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
	Token    string      `json:"token,omitempty"`
	Error    interface{} `json:"error,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

// JSONResponse sends a standard JSON response to the client.
func JSONResponse(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := GeneralResponse{
		Success: success,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// JSONError sends a standard JSON error response.
func JSONError(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, false, message, nil)
}

// JSONSuccess sends a standard JSON success response.
func JSONSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	JSONResponse(w, status, true, message, data)
}
