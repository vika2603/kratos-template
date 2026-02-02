package handler

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type Responder interface {
	Success(c *app.RequestContext, resp any)
	Fail(c *app.RequestContext, err error)
}

type DefaultResponder struct {
	logger log.Logger
}

func Default(logger log.Logger) Responder {
	return &DefaultResponder{logger: logger}
}

func (r *DefaultResponder) Success(c *app.RequestContext, resp any) {
	c.JSON(consts.StatusOK, resp)
}

func (r *DefaultResponder) Fail(c *app.RequestContext, err error) {
	se := errors.FromError(err)

	if se.Reason != "" {
		c.JSON(int(se.Code), ErrorResponse{
			Code:    se.Reason,
			Message: se.Message,
		})
		return
	}

	if r.logger != nil {
		log.NewHelper(r.logger).Errorw(
			"msg", "unexpected error",
			"error", err.Error(),
			"path", string(c.Request.URI().Path()),
			"method", string(c.Request.Method()),
		)
	}

	c.JSON(consts.StatusInternalServerError, ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "internal server error",
	})
}
