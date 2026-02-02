package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"kratos-template/app/gateway/biz/model/auth"
	"kratos-template/app/gateway/pkg/handler"
)

type AuthHandler struct {
	resp handler.Responder
	svc  auth.AuthService
}

func NewAuthHandler(resp handler.Responder, svc auth.AuthService) *AuthHandler {
	return &AuthHandler{resp: resp, svc: svc}
}

func (h *AuthHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req auth.LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.Login(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *AuthHandler) Refresh(ctx context.Context, c *app.RequestContext) {
	var req auth.RefreshRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.Refresh(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *AuthHandler) Logout(ctx context.Context, c *app.RequestContext) {
	var req auth.LogoutRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.Logout(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}
