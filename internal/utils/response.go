package utils

import (
	"encoding/json"
	"net/http"
)

// APIResponse structura standard pentru răspunsuri
type APIResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// APIError structura pentru erori
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// WriteJSON scrie un răspuns JSON
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

// WriteSuccess scrie un răspuns de succes
func WriteSuccess(w http.ResponseWriter, message string, data interface{}) {
    response := APIResponse{
        Success: true,
        Message: message,
        Data:    data,
    }
    WriteJSON(w, http.StatusOK, response)
}

// WriteCreated scrie un răspuns pentru resurse create
func WriteCreated(w http.ResponseWriter, message string, data interface{}) {
    response := APIResponse{
        Success: true,
        Message: message,
        Data:    data,
    }
    WriteJSON(w, http.StatusCreated, response)
}

// WriteError scrie un răspuns de eroare
func WriteError(w http.ResponseWriter, status int, message string, details ...string) {
    detail := ""
    if len(details) > 0 {
        detail = details[0]
    }
    
    response := APIResponse{
        Success: false,
        Error:   message,
    }
    
    if detail != "" {
        response.Data = APIError{
            Code:    status,
            Message: message,
            Details: detail,
        }
    }
    
    WriteJSON(w, status, response)
}

// WriteNoContent scrie un răspuns 204 No Content
func WriteNoContent(w http.ResponseWriter) {
    w.WriteHeader(http.StatusNoContent)
}