package validate

import (
	"context"

	"buf.build/go/protovalidate"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"google.golang.org/protobuf/proto"
)

func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			msg, ok := req.(proto.Message)
			if !ok {
				return handler(ctx, req)
			}
			if err := protovalidate.Validate(msg); err != nil {
				return nil, errors.BadRequest("VALIDATION_FAILED", err.Error())
			}
			return handler(ctx, req)
		}
	}
}
