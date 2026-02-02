package service

import (
	"errors"

	"kratos-template/app/user/internal/biz"
	pkgerrors "kratos-template/pkg/errors"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, biz.ErrUserNotFound):
		return pkgerrors.NewNotFound("USER_NOT_FOUND", "user not found")
	case errors.Is(err, biz.ErrUsernameExists):
		return pkgerrors.NewConflict("USERNAME_EXISTS", "username already exists")
	case errors.Is(err, biz.ErrEmailExists):
		return pkgerrors.NewConflict("EMAIL_EXISTS", "email already exists")
	default:
		return pkgerrors.NewInternalServer("INTERNAL_ERROR", err.Error())
	}
}
