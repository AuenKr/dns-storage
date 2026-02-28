package pkg

import (
	"dns-storage/pkg/defaults"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(defaults.NewDefaultConfig),
	fx.Provide(defaults.NewLogger),
)
