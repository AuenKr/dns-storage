package handler

import "context"

type DNSTXTProvider interface {
	CreateTXTRecord(ctx context.Context, subdomain string, value string) (Record, error)
	GetTXTRecords(ctx context.Context, subdomain string) (Record, error)
	DeleteTXTRecord(ctx context.Context, id string) error
	UpdateTXTRecord(ctx context.Context, id string, record Record) (Record, error)
}

type RecordType string

const (
	TXTRecord RecordType = "TXT"
)

type Record struct {
	ID        string     `json:"id"`
	Subdomain string     `json:"name"`
	Type      RecordType `json:"type"`
	Content   string     `json:"content"`
	TTL       int        `json:"ttl"`
	Comment   string     `json:"comment"`
}
