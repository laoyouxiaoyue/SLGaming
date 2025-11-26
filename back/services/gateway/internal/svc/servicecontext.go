// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SLGaming/back/services/gateway/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
