package asset

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/asset"
)

func (s *AssetService) GetAsset(ctx context.Context, req *asset.GetAssetRequest) (*asset.GetAssetResponse, error) {
	return nil, errors.New("not implemented")
}
