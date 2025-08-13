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

// --- Helper-e pentru status code-uri uzuale ---

// WriteAccepted 202 - pentru operații acceptate/asynchronize
func WriteAccepted(w http.ResponseWriter, message string, data interface{}) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	WriteJSON(w, http.StatusAccepted, response)
}

// Erori 4xx
func WriteBadRequest(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusBadRequest, message, details...)
}

func WriteUnauthorized(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusUnauthorized, message, details...)
}

func WriteForbidden(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusForbidden, message, details...)
}

func WriteNotFound(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusNotFound, message, details...)
}

func WriteMethodNotAllowed(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusMethodNotAllowed, message, details...)
}

// 409 Conflict - ex: email/telefon deja existent
func WriteConflict(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusConflict, message, details...)
}

// 422 Unprocessable Entity - erori de validare
func WriteUnprocessableEntity(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusUnprocessableEntity, message, details...)
}

// 429 Too Many Requests - rate limiting
func WriteTooManyRequests(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusTooManyRequests, message, details...)
}

// Erori 5xx
func WriteInternalServerError(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusInternalServerError, message, details...)
}

func WriteServiceUnavailable(w http.ResponseWriter, message string, details ...string) {
	WriteError(w, http.StatusServiceUnavailable, message, details...)
}