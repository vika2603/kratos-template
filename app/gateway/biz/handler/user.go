package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"kratos-template/app/gateway/biz/model/user"
	"kratos-template/app/gateway/pkg/handler"
)

type UserHandler struct {
	resp handler.Responder
	svc  user.UserService
}

func NewUserHandler(resp handler.Responder, svc user.UserService) *UserHandler {
	return &UserHandler{resp: resp, svc: svc}
}

func (h *UserHandler) ListUsers(ctx context.Context, c *app.RequestContext) {
	var req user.ListUsersRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.ListUsers(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *UserHandler) CreateUser(ctx context.Context, c *app.RequestContext) {
	var req user.CreateUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.CreateUser(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *UserHandler) GetUser(ctx context.Context, c *app.RequestContext) {
	var req user.GetUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.GetUser(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *UserHandler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	var req user.UpdateUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.UpdateUser(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}

func (h *UserHandler) DeleteUser(ctx context.Context, c *app.RequestContext) {
	var req user.DeleteUserRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.DeleteUser(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}
