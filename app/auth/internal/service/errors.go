package service

import (
	"errors"

	"kratos-template/app/auth/internal/biz"
	pkgerrors "kratos-template/pkg/errors"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, biz.ErrUserNotFound):
		return pkgerrors.NewNotFound("USER_NOT_FOUND", "user not found")
	case errors.Is(err, biz.ErrInvalidCredentials):
		return pkgerrors.NewUnauthorized("INVALID_CREDENTIALS", "invalid username or password")
	default:
		return pkgerrors.NewInternalServer("INTERNAL_ERROR", err.Error())
	}
}
