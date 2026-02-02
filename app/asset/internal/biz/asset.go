package biz

import (
	"context"
	"errors"

	"kratos-template/pkg/model"
)

type AssetRepo interface {
	Create(ctx context.Context, asset *model.Asset) error
	Get(ctx context.Context, id string) (*model.Asset, error)
	Update(ctx context.Context, asset *model.Asset) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, pageSize int32, ownerID string) ([]*model.Asset, int64, error)
}

var ErrAssetNotFound = errors.New("asset not found")

type AssetUseCase struct {
	repo AssetRepo
}

func (uc *AssetUseCase) Create(ctx context.Context, name, description, ownerID string, value float64) (*model.Asset, error) {
	asset := &model.Asset{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		Value:       value,
	}

	if err := uc.repo.Create(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (uc *AssetUseCase) Get(ctx context.Context, id string) (*model.Asset, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *AssetUseCase) Update(ctx context.Context, id, name, description string, value float64) (*model.Asset, error) {
	asset, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		asset.Name = name
	}
	if description != "" {
		asset.Description = description
	}
	if value > 0 {
		asset.Value = value
	}

	if err := uc.repo.Update(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (uc *AssetUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *AssetUseCase) List(ctx context.Context, page, pageSize int32, ownerID string) ([]*model.Asset, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return uc.repo.List(ctx, page, pageSize, ownerID)
}
