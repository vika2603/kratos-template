package data

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	userv1 "kratos-template/api/user/v1"
)

func TestTranslateDBError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		in   error
		want func(error) bool
	}{
		{"not found", gorm.ErrRecordNotFound, userv1.IsUserNotFound},
		{
			"duplicate username",
			&pgconn.PgError{Code: "23505", ConstraintName: "users_username_key"},
			userv1.IsUsernameExists,
		},
		{
			"duplicate email wrapped",
			fmt.Errorf("create: %w", &pgconn.PgError{Code: "23505", ConstraintName: "users_email_key"}),
			userv1.IsEmailExists,
		},
		{
			"unknown unique constraint",
			&pgconn.PgError{Code: "23505", ConstraintName: "users_something_key"},
			userv1.IsInternal,
		},
		{"arbitrary error", errors.New("connection reset"), userv1.IsInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateDBError(ctx, tt.in)
			if !tt.want(got) {
				t.Errorf("translateDBError(%v) = %v, wrong reason", tt.in, got)
			}
		})
	}

	if translateDBError(ctx, nil) != nil {
		t.Error("translateDBError(nil) != nil")
	}
}
