package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"kratos-template/app/gateway/biz/model/asset"
	"kratos-template/app/gateway/pkg/handler"
)

type AssetHandler struct {
	resp handler.Responder
	svc  asset.AssetService
}

func NewAssetHandler(resp handler.Responder, svc asset.AssetService) *AssetHandler {
	return &AssetHandler{resp: resp, svc: svc}
}

func (h *AssetHandler) ListAssets(ctx context.Context, c *app.RequestContext) {
	var req asset.ListAssetsRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.ListAssets(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *AssetHandler) CreateAsset(ctx context.Context, c *app.RequestContext) {
	var req asset.CreateAssetRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.CreateAsset(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *AssetHandler) GetAsset(ctx context.Context, c *app.RequestContext) {
	var req asset.GetAssetRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.GetAsset(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *AssetHandler) UpdateAsset(ctx context.Context, c *app.RequestContext) {
	var req asset.UpdateAssetRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.UpdateAsset(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *AssetHandler) DeleteAsset(ctx context.Context, c *app.RequestContext) {
	var req asset.DeleteAssetRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.DeleteAsset(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}
