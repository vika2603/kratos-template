package errors

type BizError struct {
	Status  int
	Code    int
	Message string
	Data    any
}

func (e *BizError) Error() string {
	return e.Message
}

func (e *BizError) WithData(data any) *BizError {
	e.Data = data
	return e
}

func New(status, code int, message string) *BizError {
	return &BizError{Status: status, Code: code, Message: message}
}

func BadRequest(message string) *BizError {
	return &BizError{Status: 400, Code: 400, Message: message}
}

func Unauthorized(message string) *BizError {
	return &BizError{Status: 401, Code: 401, Message: message}
}

func Forbidden(message string) *BizError {
	return &BizError{Status: 403, Code: 403, Message: message}
}

func NotFound(message string) *BizError {
	return &BizError{Status: 404, Code: 404, Message: message}
}

func Conflict(message string) *BizError {
	return &BizError{Status: 409, Code: 409, Message: message}
}

func Internal(message string) *BizError {
	return &BizError{Status: 500, Code: 500, Message: message}
}
