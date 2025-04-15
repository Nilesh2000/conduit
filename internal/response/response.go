package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// GenericErrorModel represents the API error response body
type GenericErrorModel struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}

// RespondWithError sends an error response with the given status code and errors
func RespondWithError(w http.ResponseWriter, status int, errors []string) {
	w.WriteHeader(status)

	response := GenericErrorModel{}
	response.Errors.Body = errors

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
