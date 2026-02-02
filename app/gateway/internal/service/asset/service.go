package asset

import (
	"kratos-template/app/gateway/biz/model/asset"
)

type AssetService struct{}

func NewService() asset.AssetService {
	return &AssetService{}
}
