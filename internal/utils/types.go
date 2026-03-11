// Package utils contains helper functions
package utils

import (
	"encoding/base64"

	"github.com/klauspost/compress/zstd"
)

// ChunkTXTRrecord : will be compressed and base64 encoded will setting value and vice versa while reading
type ChunkTXTRrecord struct {
	record string
}

func (c *ChunkTXTRrecord) SetRecord(record []byte) error {
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return err
	}
	compress := encoder.EncodeAll(record, nil)
	base64Value := base64.StdEncoding.EncodeToString(compress)
	c.record = base64Value
	return nil
}

func (c *ChunkTXTRrecord) GetRecord() ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	bytesValue, err := base64.StdEncoding.DecodeString(c.record)
	if err != nil {
		return nil, err
	}
	return decoder.DecodeAll(bytesValue, nil)
}
