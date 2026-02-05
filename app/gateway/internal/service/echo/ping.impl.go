package echo

import (
	"context"

	"kratos-template/app/gateway/biz/model/echo"

	"github.com/go-kratos/kratos/v2/errors"
)

func (s *EchoService) Ping(ctx context.Context, req *echo.PingRequest) (*echo.PingResponse, error) {
	return nil, errors.New(400, "BAD_REQUEST", "fuck")
}
