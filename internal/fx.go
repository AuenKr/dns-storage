package internal

import (
	"dns-storage/internal/client"
	"dns-storage/internal/handler"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(handler.NewDNSHandler),
	fx.Provide(handler.NewCloudflareDNSProviderClient),
	fx.Provide(handler.NewBunnyDNSProvider),
	fx.Provide(handler.NewFileHander),
	fx.Provide(client.NewDNSClient),
)
