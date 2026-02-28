package defaults

import (
	"fmt"
	"os"
	"strconv"
)

type DefaultConfig struct {
	Domain           string
	TTL              int // in sec
	DNSServerAddress string
	DownloadDir      string

	CloudflareBaseURL  string
	CloudflareAPIToken string
	CloudflareZoneID   string

	BunnyBaseURL  string
	BunnyAPIToken string
	BunnyZoneID   string
}

func NewDefaultConfig() *DefaultConfig {
	domain := getDefaultValue("DOMAIN", "auenkr.qzz.io")
	cloudflareBaseURL := getDefaultValue("CLOUDFLARE_BASE_URL", "https://api.cloudflare.com/client/v4")

	cloudflareAPIToken := getDefaultValue("CLOUDFLARE_API_TOKEN", "")
	cloudflareZoneID := getDefaultValue("CLOUDFLARE_ZONE_ID", "")

	dnsServerAddress := getDefaultValue("DNS_SERVER_ADDRESS", "1.1.1.1:53")
	downloadDir := getDefaultValue("DOWNLOAD_DIR", ".temp/download")

	bunnyBaseURL := getDefaultValue("BUNNY_BASE_URL", "https://api.bunny.net")
	bunnyAPIToken := getDefaultValue("BUNNY_API_TOKEN", "")
	bunnyZoneID := getDefaultValue("BUNNY_ZONE_ID", "")

	ttl, _ := strconv.Atoi(getDefaultValue("TTL", "3600"))

	config := &DefaultConfig{
		Domain:             domain,
		TTL:                ttl,
		DNSServerAddress:   dnsServerAddress,
		DownloadDir:        downloadDir,
		BunnyBaseURL:       bunnyBaseURL,
		BunnyAPIToken:      bunnyAPIToken,
		BunnyZoneID:        bunnyZoneID,
		CloudflareBaseURL:  cloudflareBaseURL,
		CloudflareAPIToken: cloudflareAPIToken,
		CloudflareZoneID:   cloudflareZoneID,
	}

	fmt.Println("Config:", config)
	return config
}

func getDefaultValue(env string, defaultValue string) string {
	res := os.Getenv(env)
	if res == "" {
		res = defaultValue
	}
	return res
}
