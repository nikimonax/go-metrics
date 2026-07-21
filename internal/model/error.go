package model

type ErrorResponse struct {
	Errors map[string]string `json:"errors"`
}

func NewErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Errors: make(map[string]string),
	}
}
