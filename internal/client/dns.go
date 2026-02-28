package client

import "codeberg.org/miekg/dns"

func NewDNSClient() *dns.Client {
	return dns.NewClient()
}
