package handler

import (
	"context"
	"fmt"
	"strings"

	"dns-storage/pkg/defaults"

	"codeberg.org/miekg/dns"
)

type DNSTXTHandler interface {
	ReadTXTRecord(context.Context, string) (string, error)
}

type DNSHandler struct {
	config *defaults.DefaultConfig
	client *dns.Client
}

func NewDNSHandler(client *dns.Client, config *defaults.DefaultConfig) DNSTXTHandler {
	return &DNSHandler{
		config: config,
		client: client,
	}
}

func (d *DNSHandler) ReadTXTRecord(ctx context.Context, record string) (string, error) {
	fmt.Println("TXT Record for : ", record)
	msg := dns.NewMsg(record, dns.TypeTXT)
	resp, _, err := d.client.Exchange(ctx, msg, "udp", d.config.DNSServerAddress)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("nil DNS response")
	}
	if resp.Rcode != dns.RcodeSuccess {
		return "", fmt.Errorf("dns query failed: %s", dns.RcodeToString[resp.Rcode])
	}
	for _, rr := range resp.Answer {
		txt, ok := rr.(*dns.TXT)
		if !ok {
			continue
		}
		if len(txt.Txt) == 0 {
			return "", fmt.Errorf("empty TXT answer")
		}
		return strings.Join(txt.Txt, ""), nil
	}
	return "", fmt.Errorf("no TXT record found for %q", record)
}
