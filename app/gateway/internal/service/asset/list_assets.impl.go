package asset

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/asset"
)

func (s *AssetService) ListAssets(ctx context.Context, req *asset.ListAssetsRequest) (*asset.ListAssetsResponse, error) {
	return nil, errors.New("not implemented")
}
