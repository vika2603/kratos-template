package echo

import (
	"context"

	"kratos-template/app/gateway/biz/model/echo"
)

func (s *EchoService) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoResponse, error) {
	return &echo.EchoResponse{Message: "hello: " + req.Message}, nil
}
