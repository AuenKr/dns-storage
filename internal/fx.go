package internal

import (
	"dns-storage/internal/cli"
	"dns-storage/internal/client"
	"dns-storage/internal/handler"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(handler.NewDNSHandler),
	fx.Provide(handler.NewBunnyDNSProvider),
	fx.Provide(handler.NewFileHander),
	fx.Provide(client.NewDNSClient),
	fx.Provide(cli.NewFlags),
	fx.Provide(cli.NewCommandLine),
)
