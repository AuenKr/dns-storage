package client

import (
	"dns-storage/pkg/defaults"

	"github.com/cloudflare/cloudflare-go/v3"
	"github.com/cloudflare/cloudflare-go/v3/option"
)

func NewCloudflareClient(config *defaults.DefaultConfig) *cloudflare.Client {
	client := cloudflare.NewClient(option.WithAPIToken(config.Token))
	return client
}
