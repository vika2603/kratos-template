package data

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"kratos-template/app/asset/internal/biz"
	"kratos-template/pkg/model"
)

var _ biz.AssetRepo = (*AssetRepo)(nil)

type AssetRepo struct {
	data *Data
}

func (r *AssetRepo) Create(ctx context.Context, asset *model.Asset) error {
	return r.data.q.Asset.WithContext(ctx).Create(asset)
}

func (r *AssetRepo) Get(ctx context.Context, id string) (*model.Asset, error) {
	asset, err := r.data.q.Asset.WithContext(ctx).Where(r.data.q.Asset.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrAssetNotFound
		}
		return nil, err
	}
	return asset, nil
}

func (r *AssetRepo) Update(ctx context.Context, asset *model.Asset) error {
	_, err := r.data.q.Asset.WithContext(ctx).Where(r.data.q.Asset.ID.Eq(asset.ID)).Updates(asset)
	return err
}

func (r *AssetRepo) Delete(ctx context.Context, id string) error {
	_, err := r.data.q.Asset.WithContext(ctx).Where(r.data.q.Asset.ID.Eq(id)).Delete()
	return err
}

func (r *AssetRepo) List(ctx context.Context, page, pageSize int32, ownerID string) ([]*model.Asset, int64, error) {
	q := r.data.q.Asset.WithContext(ctx)

	if ownerID != "" {
		q = q.Where(r.data.q.Asset.OwnerID.Eq(ownerID))
	}

	total, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	offset := int((page - 1) * pageSize)
	assets, err := q.Limit(int(pageSize)).Offset(offset).Find()
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}
