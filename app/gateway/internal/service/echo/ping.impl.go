package echo

import (
	"context"

	"kratos-template/app/gateway/biz/model/echo"
	"kratos-template/app/gateway/pkg/errors"
)

func (s *EchoService) Ping(ctx context.Context, req *echo.PingRequest) (*echo.PingResponse, error) {
	return nil, errors.BadRequest("fuck").WithData("fuck you data")
}
