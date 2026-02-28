package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"dns-storage/pkg/defaults"

	"codeberg.org/miekg/dns"
)

type DNSTXTHandler interface {
	ReadTXTRecord(context.Context, string) ([]byte, error)
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

func (d *DNSHandler) ReadTXTRecord(ctx context.Context, record string) ([]byte, error) {
	msg := dns.NewMsg(record, dns.TypeTXT)
	ctx = context.TODO()
	resp, _, err := d.client.Exchange(ctx, msg, "udp", d.config.DNSServerAddress)
	if err != nil {
		fmt.Println("err", err)
		fmt.Printf("%#v\n", resp)
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("nil DNS response")
	}
	if resp.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("dns query failed: %s", dns.RcodeToString[resp.Rcode])
	}
	for _, rr := range resp.Answer {
		txt, ok := rr.(*dns.TXT)
		if !ok {
			continue
		}
		if len(txt.Txt) == 0 {
			return nil, fmt.Errorf("empty TXT answer")
		}
		return base64.StdEncoding.DecodeString(strings.Join(txt.Txt, ""))
	}
	return nil, fmt.Errorf("no TXT record found for %q", record)
}
