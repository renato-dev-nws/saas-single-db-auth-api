package shared

// PaginatedResponse is the standard paginated response format
type PaginatedResponse struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error string `json:"error"`
}

// ValidationErrorResponse holds validation errors
type ValidationErrorResponse struct {
	Errors map[string]string `json:"errors"`
}

// MessageResponse is a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
