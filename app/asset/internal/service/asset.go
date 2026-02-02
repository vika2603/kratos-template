package service

import (
	"context"

	"go.uber.org/fx"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "kratos-template/api/asset/v1"
	"kratos-template/app/asset/internal/biz"
	"kratos-template/pkg/model"
)

type AssetService struct {
	v1.UnimplementedAssetServiceServer
	assetUC *biz.AssetUseCase
}

type AssetServiceParams struct {
	fx.In
	AssetUseCase *biz.AssetUseCase
}

type AssetServiceResult struct {
	fx.Out
	AssetService v1.AssetServiceServer
}

func NewAssetService(params AssetServiceParams) AssetServiceResult {
	return AssetServiceResult{
		AssetService: &AssetService{
			assetUC: params.AssetUseCase,
		},
	}
}

func (s *AssetService) CreateAsset(ctx context.Context, req *v1.CreateAssetRequest) (*v1.CreateAssetReply, error) {
	asset, err := s.assetUC.Create(ctx, req.Name, req.Description, req.OwnerId, req.Value)
	if err != nil {
		return nil, mapError(err)
	}

	return &v1.CreateAssetReply{
		Asset: modelToProto(asset),
	}, nil
}

func (s *AssetService) GetAsset(ctx context.Context, req *v1.GetAssetRequest) (*v1.GetAssetReply, error) {
	asset, err := s.assetUC.Get(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}

	return &v1.GetAssetReply{
		Asset: modelToProto(asset),
	}, nil
}

func (s *AssetService) UpdateAsset(ctx context.Context, req *v1.UpdateAssetRequest) (*v1.UpdateAssetReply, error) {
	asset, err := s.assetUC.Update(ctx, req.Id, req.Name, req.Description, req.Value)
	if err != nil {
		return nil, mapError(err)
	}

	return &v1.UpdateAssetReply{
		Asset: modelToProto(asset),
	}, nil
}

func (s *AssetService) DeleteAsset(ctx context.Context, req *v1.DeleteAssetRequest) (*v1.DeleteAssetReply, error) {
	if err := s.assetUC.Delete(ctx, req.Id); err != nil {
		return nil, mapError(err)
	}

	return &v1.DeleteAssetReply{
		Success: true,
	}, nil
}

func (s *AssetService) ListAssets(ctx context.Context, req *v1.ListAssetsRequest) (*v1.ListAssetsReply, error) {
	assets, total, err := s.assetUC.List(ctx, req.Page, req.PageSize, req.OwnerId)
	if err != nil {
		return nil, mapError(err)
	}

	protoAssets := make([]*v1.Asset, 0, len(assets))
	for _, asset := range assets {
		protoAssets = append(protoAssets, modelToProto(asset))
	}

	return &v1.ListAssetsReply{
		Assets: protoAssets,
		Total:  int32(total),
	}, nil
}

func modelToProto(asset *model.Asset) *v1.Asset {
	return &v1.Asset{
		Id:          asset.ID,
		Name:        asset.Name,
		Description: asset.Description,
		OwnerId:     asset.OwnerID,
		Value:       asset.Value,
		CreatedAt:   timestamppb.New(asset.CreatedAt),
		UpdatedAt:   timestamppb.New(asset.UpdatedAt),
	}
}

var Module = fx.Module("service",
	fx.Provide(NewAssetService),
)
