package handler

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/go-kratos/kratos/v2/errors"
	"go.uber.org/zap"

	"kratos-template/pkg/log"
)

type Responder interface {
	Success(c *app.RequestContext, resp any)
	Fail(c *app.RequestContext, err error)
}

type DefaultResponder struct{}

func Default() Responder {
	return &DefaultResponder{}
}

func (r *DefaultResponder) Success(c *app.RequestContext, resp any) {
	c.JSON(consts.StatusOK, resp)
}

func (r *DefaultResponder) Fail(c *app.RequestContext, err error) {
	if se := errors.FromError(err); se.Reason != "" {
		c.JSON(int(se.Code), ErrorResponse{
			Code:    se.Reason,
			Message: se.Message,
		})
		return
	}

	log.Error("unexpected error",
		zap.String("error", err.Error()),
		zap.String("path", string(c.Request.URI().Path())),
		zap.String("method", string(c.Request.Method())),
	)

	c.JSON(consts.StatusInternalServerError, ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "internal server error",
	})
}
