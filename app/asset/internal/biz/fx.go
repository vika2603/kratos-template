package biz

import "go.uber.org/fx"

func NewAssetUseCase(repo AssetRepo) *AssetUseCase {
	return &AssetUseCase{repo: repo}
}

var Module = fx.Module("asset.biz",
	fx.Provide(NewAssetUseCase),
)
