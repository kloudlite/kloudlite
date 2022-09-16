package app

import (
	"go.uber.org/fx"
	"kloudlite.io/apps/nodectrl/internal/domain"
)

var Module = fx.Module(
	"app",
	domain.Module,
)
