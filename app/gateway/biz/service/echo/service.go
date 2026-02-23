package echo

import "kratos-template/app/gateway/biz/model/echo"

type EchoService struct{}

func NewService() echo.EchoService {
	return &EchoService{}
}
