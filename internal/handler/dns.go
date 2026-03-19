package handler

import (
	"context"
	"fmt"
	"strings"

	"dns-storage/pkg/defaults"

	"codeberg.org/miekg/dns"
)

type DNSTXTHandler interface {
	ReadTXTRecord(domain string) (string, error)
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

func (d *DNSHandler) ReadTXTRecord(domain string) (string, error) {
	msg := dns.NewMsg(domain, dns.TypeTXT)
	resp, _, err := d.client.Exchange(context.TODO(), msg, string(d.config.NetworkLayer), d.config.DNSServerAddress)
	if err != nil {
		fmt.Println("err", err)
		fmt.Printf("%#v\n", resp)
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("nil DNS response")
	}
	if resp.Rcode != dns.RcodeSuccess {
		// NXDOMAIN means: the DNS server replied successfully but the queried name does not exist in DNS
		return "", fmt.Errorf("dns query failed: %s\n%#v", dns.RcodeToString[resp.Rcode], resp)
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
	return "", fmt.Errorf("no TXT record found for %q %#v", domain, resp)
}
