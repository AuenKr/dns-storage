package handler

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dns-storage/pkg/defaults"
)

// MaxChunkSize bytes store in TXT is 255 character limit
// Storing record into base64 format: it increase character
// After doing maths : N <= 191
// const MaxChunkSize = 255
const MaxChunkSize = 191

type FileHandler interface {
	Upload(ctx context.Context, filePath string) (string, error)          // return index file record
	Download(ctx context.Context, indexFileRecord string) (string, error) // return file path
	Delete(ctx context.Context, record string) error
}

type FileUploader struct {
	config            *defaults.DefaultConfig
	dnsProviderClient *CloudflareDNS
	dnsClient         DNSTXTHandler
}

func NewFileHander(
	config *defaults.DefaultConfig,
	dnsProvierCli *CloudflareDNS,
	dnsCli DNSTXTHandler,
) FileHandler {
	return &FileUploader{
		config:            config,
		dnsProviderClient: dnsProvierCli,
		dnsClient:         dnsCli,
	}
}

type FileHandlerUploadResponse struct {
	FilePath string `json:"file_path"` // txt record of index
}

func (f *FileUploader) Upload(ctx context.Context, filePath string) (string, error) {
	// Canculate the no of chunks need to create
	fmt.Println("FilePath:", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	name := fileInfo.Name()
	fmt.Println("TotalChunks", math.Ceil(float64(fileInfo.Size())/float64(MaxChunkSize)))

	// Start reading the file into MAX_CHUNK_SIZE byte and add its txt record (io.pipe)
	data := make([]byte, MaxChunkSize)
	noChunks := 0
	var offset int64 = 0
	for {
		n, err := file.ReadAt(data, offset)
		if n == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		data = data[:n]
		subdomain := fmt.Sprintf("%s.%d.%s", name, noChunks, f.config.BaseURL)
		_, err = f.dnsProviderClient.CreateTXTRecord(ctx, subdomain, data)
		if err != nil {
			return "", err
		}
		offset += int64(n)
		noChunks++
		fmt.Println("File Chunk", noChunks, subdomain)
		time.Sleep(50 * time.Millisecond)
	}

	// Create a index TXT record for that chuck file
	// <NAME.FILE_TYPE>.<END_CHUNK|100>.<DomainName>
	// TXT Records: <NAME.FILE_TYPE>.<END_CHUNK_NO>
	record := fmt.Sprintf("%s.%s", name, f.config.BaseURL)
	txtRecord := []byte(fmt.Sprintf("%s.%d", name, noChunks))

	createIndexRecord, err := f.dnsProviderClient.CreateTXTRecord(ctx, record, txtRecord)
	if err != nil {
		return "", err
	}

	return createIndexRecord.Result.Name, nil
}

func (f *FileUploader) Download(ctx context.Context, indexFileRecord string) (string, error) {
	// Get TXT Record
	txtRecord, err := f.dnsClient.ReadTXTRecord(ctx, indexFileRecord)
	if err != nil {
		return "", err
	}

	// TXT Records: <NAME.FILE_TYPE>.<END_CHUNK|100>
	temp := strings.Split(string(txtRecord), ".")
	fileName := strings.Join(temp[:len(temp)-1], ".")
	noChunks, err := strconv.Atoi(temp[len(temp)-1])
	if err != nil {
		return "", err
	}
	fmt.Println("FileName:", fileName)
	fmt.Println("Total noChunks:", noChunks)

	downloadPath := filepath.Join(f.config.DownloadDir, fileName)
	file, err := os.Create(downloadPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	for i := range noChunks {
		domain := fmt.Sprintf("%s.%d.%s", fileName, i, f.config.BaseURL)
		txtRecord, err := f.dnsClient.ReadTXTRecord(ctx, domain)
		if err != nil {
			return "", err
		}
		file.WriteAt(txtRecord, int64(i)*MaxChunkSize)
		fmt.Println("File Chunk", i)
		time.Sleep(20 * time.Millisecond)
	}
	return downloadPath, nil
}

func (f *FileUploader) Delete(ctx context.Context, indexFile string) error {
	// Check correct index file record
	// Get TXT Record
	txtRecord, err := f.dnsClient.ReadTXTRecord(ctx, indexFile)
	if err != nil {
		return err
	}

	fmt.Println("TXT Record:", string(txtRecord))
	temp := strings.Split(string(txtRecord), ".")
	fileName := strings.Join(temp[:len(temp)-1], ".")
	noChunks, err := strconv.Atoi(temp[len(temp)-1])
	if err != nil {
		return err
	}

	for i := range noChunks {
		subdomain := fmt.Sprintf("%s.%d.%s", fileName, i, f.config.BaseURL)
		record, err := f.dnsProviderClient.GetTXTRecords(ctx, subdomain)
		if err != nil {
			return err
		}
		_, err = f.dnsProviderClient.DeleteTXTRecord(ctx, record.ID)
		if err != nil {
			return err
		}
		fmt.Println("Deleted", subdomain)
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}

func (f *FileUploader) Stream(ctx context.Context, record string) error {
	// I do not know it
	panic("implement me")
	return nil
}
