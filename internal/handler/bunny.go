package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dns-storage/pkg/defaults"
)

type BunnyDNSProvider struct {
	config *defaults.DefaultConfig
}

func NewBunnyDNSProvider(
	config *defaults.DefaultConfig,
) DNSTXTProvider {
	return &BunnyDNSProvider{
		config: config,
	}
}

type (
	BunnyCreateRecordRequest struct {
		ID      int    `json:"Id"`
		Name    string `json:"Name"`
		Type    int    `json:"Type"`
		TTL     int    `json:"Ttl"` // TXT = 3
		Value   string `json:"Value"`
		Comment string `json:"Comment"`
	}
	BunnyDeleteRecordResponse = BunnyCreateRecordRequest
)

// CreateTXTRecord implements [DNSTXTProvider].
func (b *BunnyDNSProvider) CreateTXTRecord(ctx context.Context, subdomain string, value string) (Record, error) {
	if len(value) > b.config.MaxTXTRecordCharacterSize {
		return Record{}, fmt.Errorf("txt record value is too long")
	}

	url := fmt.Sprintf("%s/dnszone/%s/records", b.config.BunnyBaseURL, b.config.BunnyZoneID)

	bunnyRecord := BunnyCreateRecordRequest{
		Name:    subdomain,
		Type:    3,
		TTL:     b.config.TTL,
		Value:   value,
		Comment: b.config.Comment,
	}

	payload, err := json.Marshal(bunnyRecord)
	if err != nil {
		fmt.Println("Error while unmarshalling bunny record")
		return Record{}, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(payload))
	if err != nil {
		fmt.Println("Error while creating req")
		return Record{}, err
	}
	req.Header.Add("AccessKey", b.config.BunnyAPIToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error while doing req")
		return Record{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Record{}, err
	}
	var result BunnyDeleteRecordResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println(string(body))
		return Record{}, err
	}

	return Record{
		ID:        result.ID,
		Subdomain: result.Name,
		Type:      TXTRecord,
		Content:   result.Value,
		TTL:       result.TTL,
		Comment:   result.Comment,
	}, nil
}

type BunnyGetZoneResponse struct {
	ID          int                         `json:"Id"`
	Domain      string                      `json:"Domain"`
	Records     []BunnyDeleteRecordResponse `json:"Records"`
	Nameserver1 string                      `json:"Nameserver1"`
	Nameserver2 string                      `json:"Nameserver2"`
	SoaEmail    string                      `json:"SoaEmail"`
}

var (
	CacheGetResponse BunnyGetZoneResponse
	lastCacheTime    time.Time = time.Now().Add(-5 * time.Minute)
)

// GetTXTRecords implements [DNSTXTProvider].
func (b *BunnyDNSProvider) GetTXTRecords(ctx context.Context, subdomain string) (Record, error) {
	records, err := b.GetAllRecord(ctx)
	if err != nil {
		return Record{}, err
	}

	for _, value := range records {
		if value.Subdomain == subdomain {
			return value, nil
		}
	}
	return Record{}, fmt.Errorf("%s not found in record", subdomain)
}

// DeleteTXTRecord implements [DNSTXTProvider].
func (b *BunnyDNSProvider) DeleteTXTRecord(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/dnszone/%s/records/%s", b.config.BunnyBaseURL, b.config.BunnyZoneID, id)

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Add("AccessKey", b.config.BunnyAPIToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("status code: %d, %s", res.StatusCode, string(body))
	}
	return nil
}

// UpdateTXTRecord implements [DNSTXTProvider].
func (b *BunnyDNSProvider) UpdateTXTRecord(ctx context.Context, id string, record Record) (Record, error) {
	if len(record.Content) > b.config.MaxTXTRecordCharacterSize {
		return Record{}, fmt.Errorf("txt record value is too long")
	}
	url := fmt.Sprintf("%s/dnszone/%s/records/%s", b.config.BunnyBaseURL, b.config.BunnyZoneID, id)

	bunnyRecord := BunnyCreateRecordRequest{
		Name:    record.Subdomain,
		Type:    3,
		Value:   record.Content,
		TTL:     b.config.TTL,
		Comment: b.config.Comment,
	}
	payload, err := json.Marshal(bunnyRecord)
	if err != nil {
		return Record{}, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return Record{}, err
	}

	req.Header.Add("AccessKey", b.config.BunnyAPIToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Record{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Record{}, err
	}
	fmt.Println(string(body))
	return record, nil
}

// GetAllRecord implements [DNSTXTProvider].
func (b *BunnyDNSProvider) GetAllRecord(ctx context.Context) ([]Record, error) {
	records := make([]Record, 0)
	url := fmt.Sprintf("%s/dnszone/%s/", b.config.BunnyBaseURL, b.config.BunnyZoneID)

	now := time.Now()
	if (lastCacheTime.Add(time.Second * time.Duration(b.config.ResponseCacheTime))).Before(now) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return records, err
		}
		req.Header.Add("AccessKey", b.config.BunnyAPIToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return records, err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return records, err
		}
		var result BunnyGetZoneResponse
		err = json.Unmarshal(body, &result)
		if err != nil {
			return records, err
		}
		CacheGetResponse = result
	}

	for _, record := range CacheGetResponse.Records {
		// 3 -> TXT record type in bunny
		if record.Type == 3 {
			records = append(records, Record{
				ID:        record.ID,
				Subdomain: record.Name,
				Type:      TXTRecord,
				Content:   record.Value,
				TTL:       record.TTL,
				Comment:   record.Comment,
			})
		}
	}

	fmt.Println("Total Records: ", len(CacheGetResponse.Records))
	fmt.Println("Total TXT Records: ", len(records))
	return records, nil
}
