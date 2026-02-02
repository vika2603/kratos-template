package asset

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/asset"
)

func (s *AssetService) DeleteAsset(ctx context.Context, req *asset.DeleteAssetRequest) (*asset.DeleteAssetResponse, error) {
	return nil, errors.New("not implemented")
}
