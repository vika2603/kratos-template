package service

import (
	"errors"

	"kratos-template/app/asset/internal/biz"
	pkgerrors "kratos-template/pkg/errors"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, biz.ErrAssetNotFound):
		return pkgerrors.NewNotFound("ASSET_NOT_FOUND", "asset not found")
	default:
		return pkgerrors.NewInternalServer("INTERNAL_ERROR", err.Error())
	}
}
