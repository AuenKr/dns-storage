package defaults

import (
	"fmt"
	"os"
)

type DefaultConfig struct {
	BaseURL           string
	CloudflareBaseURL string
	TTL               int // in sec
	DNSServerAddress  string
	Token             string
	ZoneID            string
	DownloadDir       string
}

func NewDefaultConfig() *DefaultConfig {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "auenkr.qzz.io"
	}
	cloudflareBaseURL := os.Getenv("CLOUDFLARE_BASE_URL")
	if cloudflareBaseURL == "" {
		cloudflareBaseURL = "https://api.cloudflare.com/client/v4"
	}
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	dnsServerAddress := os.Getenv("DNS_SERVER_ADDRESS")
	if dnsServerAddress == "" {
		dnsServerAddress = "1.1.1.1:53"
	}

	downloadDir := os.Getenv("DOWNLOAD_DIR")
	if downloadDir == "" {
		downloadDir = ".temp/download"
	}

	config := &DefaultConfig{
		BaseURL:           baseURL,
		CloudflareBaseURL: cloudflareBaseURL,
		TTL:               1,
		Token:             token,
		ZoneID:            zoneID,
		DNSServerAddress:  dnsServerAddress,
		DownloadDir:       downloadDir,
	}

	fmt.Println("Config:", config)
	return config
}
