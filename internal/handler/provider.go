package handler

import (
	"context"
)

type DNSTXTProvider interface {
	CreateTXTRecord(ctx context.Context, subdomain string, value string) (Record, error)
	GetTXTRecords(ctx context.Context, subdomain string) (Record, error)
	DeleteTXTRecord(ctx context.Context, id string) error
	UpdateTXTRecord(ctx context.Context, id string, record Record) (Record, error)
	GetAllRecord(ctx context.Context) ([]Record, error)
}

type RecordType string

const (
	TXTRecord RecordType = "TXT"
)

type Record struct {
	ID        int        `json:"id"`
	Subdomain string     `json:"name"`
	Type      RecordType `json:"type"`
	Content   string     `json:"content"`
	TTL       int        `json:"ttl"`
	Comment   string     `json:"comment"`
}

type FileHandlerProvider interface {
	Upload(ctx context.Context, filePath string, subdomain string) (<-chan FileStatus, <-chan error)         // return Upload Chunk Status Channel
	Download(ctx context.Context, indexFileRecord string, filePath string) (<-chan FileStatus, <-chan error) // return Download Chunk Status Channel
	Delete(ctx context.Context, indexFileRecord string) (<-chan FileStatus, <-chan error)                    // return Delete Chunk Status Channel
	Stream(ctx context.Context, indexFileRecord string) (<-chan FileStream, <-chan error)                    // return Stream Chunk data
}

type FileStatus struct {
	TotalChunks  int
	CurrentChunk int
	Subdomain    string
}

type FileStream struct {
	Data     []byte
	MetaData FileStatus
}
