package echo

import (
	"context"
	"errors"

	"kratos-template/app/gateway/biz/model/echo"
)

func (s *EchoService) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoResponse, error) {
	return nil, errors.New("not implemented")
}
