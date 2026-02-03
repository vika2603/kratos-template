package grpcerr

import (
	"context"

	kratosErrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http/status"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpcstatus "google.golang.org/grpc/status"
)

func EncodeServer() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			reply, err := handler(ctx, req)
			if err == nil {
				return reply, nil
			}

			if _, ok := err.(*kratosErrors.Error); ok {
				return reply, err
			}

			ke := kratosErrors.FromError(err)
			st, _ := grpcstatus.New(status.ToGRPCCode(int(ke.Code)), ke.Message).WithDetails(
				&errdetails.ErrorInfo{Reason: ke.Reason, Metadata: ke.Metadata},
			)

			return reply, st.Err()
		}
	}
}
