package handler

// ErrorResponse is the standard error response format for the gateway.
// Business errors (with Reason set) are exposed to clients.
// Unexpected errors return generic message without internal details.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
