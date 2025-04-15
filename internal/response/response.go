package response

// GenericErrorModel represents the API error response body
type GenericErrorModel struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}
