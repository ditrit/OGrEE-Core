package models

// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error Response Message
	// in: message
	Message string `json:"message"`
}

// swagger:model SuccessResponse
type SuccessResponse struct {
	// Error Response Message
	// in: message
	Message string `json:"message"`
}


type Message struct {

	StatusCode int `json:"statuscode"`
	Message string `json:"message"`
}
