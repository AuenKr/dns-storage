package handler

type DNSProvider interface {
	CreateTXTRecord(record string, value []byte) (CreateRecordResponse, error)
	DeleteTXTRecord(id string) (DeleteRecordResponse, error)
}
