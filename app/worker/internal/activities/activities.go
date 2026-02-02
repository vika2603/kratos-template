package activities

import (
	"github.com/go-kratos/kratos/v2/log"
)

type Activities struct {
	logger log.Logger
}

func NewActivities(logger log.Logger) *Activities {
	return &Activities{
		logger: logger,
	}
}
