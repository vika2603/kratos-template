package errors

import (
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrorUnauthorized       = errors.Unauthorized("UNAUTHORIZED", "unauthorized access")
	ErrorForbidden          = errors.Forbidden("FORBIDDEN", "forbidden access")
	ErrorNotFound           = errors.NotFound("NOT_FOUND", "resource not found")
	ErrorInternalServer     = errors.InternalServer("INTERNAL_SERVER", "internal server error")
	ErrorBadRequest         = errors.BadRequest("BAD_REQUEST", "bad request")
	ErrorValidation         = errors.BadRequest("VALIDATION_ERROR", "validation failed")
	ErrorConflict           = errors.Conflict("CONFLICT", "resource conflict")
	ErrorTooManyRequests    = errors.New(429, "TOO_MANY_REQUESTS", "too many requests")
	ErrorServiceUnavailable = errors.ServiceUnavailable("SERVICE_UNAVAILABLE", "service unavailable")
)

func NewUnauthorized(reason, message string) *errors.Error {
	return errors.Unauthorized(reason, message)
}

func NewForbidden(reason, message string) *errors.Error {
	return errors.Forbidden(reason, message)
}

func NewNotFound(reason, message string) *errors.Error {
	return errors.NotFound(reason, message)
}

func NewInternalServer(reason, message string) *errors.Error {
	return errors.InternalServer(reason, message)
}

func NewBadRequest(reason, message string) *errors.Error {
	return errors.BadRequest(reason, message)
}

func NewConflict(reason, message string) *errors.Error {
	return errors.Conflict(reason, message)
}

func NewTooManyRequests(reason, message string) *errors.Error {
	return errors.New(429, reason, message)
}

func NewServiceUnavailable(reason, message string) *errors.Error {
	return errors.ServiceUnavailable(reason, message)
}

func IsUnauthorized(err error) bool {
	return errors.IsUnauthorized(err)
}

func IsForbidden(err error) bool {
	return errors.IsForbidden(err)
}

func IsNotFound(err error) bool {
	return errors.IsNotFound(err)
}

func IsBadRequest(err error) bool {
	return errors.IsBadRequest(err)
}

func IsConflict(err error) bool {
	return errors.IsConflict(err)
}

func FromError(err error) *errors.Error {
	return errors.FromError(err)
}
