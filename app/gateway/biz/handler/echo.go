package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	"kratos-template/app/gateway/biz/model/echo"
	"kratos-template/app/gateway/pkg/handler"
)

type EchoHandler struct {
	resp handler.Responder
	svc  echo.EchoService
}

func NewEchoHandler(resp handler.Responder, svc echo.EchoService) *EchoHandler {
	return &EchoHandler{resp: resp, svc: svc}
}

func (h *EchoHandler) Echo(ctx context.Context, c *app.RequestContext) {
	var req echo.EchoRequest
	if err := c.BindAndValidate(&req); err != nil {
		h.resp.Fail(c, err)
		return
	}

	resp, err := h.svc.Echo(ctx, &req)
	if err != nil {
		h.resp.Fail(c, err)
		return
	}

	h.resp.Success(c, resp)
}
