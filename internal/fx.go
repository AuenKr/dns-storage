package internal

import (
	"dns-storage/internal/client"
	"dns-storage/internal/handler"

	"codeberg.org/miekg/dns"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(client.NewCloudflareClient),
	fx.Provide(handler.NewDNSHandler),
	fx.Provide(handler.NewDNSClient),
	fx.Provide(func() *dns.Client {
		return dns.NewClient()
	}),
)
