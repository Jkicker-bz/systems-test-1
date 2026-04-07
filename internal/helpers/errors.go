package helpers

import (
	"log"
	"net/http"
)

// ServerError logs the error and returns a 500 JSON response.
func ServerError(w http.ResponseWriter, err error) {
	log.Printf("ERROR: %v", err)
	WriteJSON(w, http.StatusInternalServerError, Envelope{
		"error": "the server encountered a problem and could not process your request",
	}, nil)
}

// NotFound returns a 404 JSON response.
func NotFound(w http.ResponseWriter) {
	WriteJSON(w, http.StatusNotFound, Envelope{
		"error": "the requested resource could not be found",
	}, nil)
}

// BadRequest returns a 400 JSON response with the provided message.
func BadRequest(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusBadRequest, Envelope{"error": msg}, nil)
}

// FailedValidation returns a 422 JSON response with field-level errors.
func FailedValidation(w http.ResponseWriter, errors map[string]string) {
	WriteJSON(w, http.StatusUnprocessableEntity, Envelope{"errors": errors}, nil)
}
