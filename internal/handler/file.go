package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dns-storage/pkg/defaults"

	"github.com/google/uuid"
)

// MaxChunkSize bytes store in TXT is 255 character limit
// Storing record into base64 format: it increase character
// After doing maths : N <= 191
// const MaxChunkSize = 255
const MaxChunkSize = 191

type FileHandler interface {
	Upload(ctx context.Context, filePath string) (string, error)          // return index file record
	Download(ctx context.Context, indexFileRecord string) (string, error) // return downloaded file path
	Delete(ctx context.Context, indexFileRecord string) error
}

type FileUploader struct {
	config            *defaults.DefaultConfig
	dnsProviderClient DNSTXTProvider
	dnsClient         DNSTXTHandler
}

func NewFileHander(
	config *defaults.DefaultConfig,
	dnsProvierCli DNSTXTProvider,
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

	startTime := time.Now()
	subdomain := uuid.New().String()
	name := fileInfo.Name()
	totalChunks := int(math.Ceil(float64(fileInfo.Size()) / float64(MaxChunkSize)))
	fmt.Println("Total Chunks:", totalChunks)

	// Create a index TXT record for that chuck file
	// TXT Records: <NAME.FILE_TYPE>.<END_CHUNK_NO>
	txtRecord := fmt.Sprintf("%s.%d", name, totalChunks)
	createIndexRecord, err := f.dnsProviderClient.CreateTXTRecord(ctx, subdomain, txtRecord)
	if err != nil {
		return "", err
	}

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
		subdomain := fmt.Sprintf("%s.%d", subdomain, noChunks)
		base64Value := base64.StdEncoding.EncodeToString(data)
		_, err = f.dnsProviderClient.CreateTXTRecord(ctx, subdomain, base64Value)
		if err != nil {
			return "", err
		}
		offset += int64(n)
		noChunks++
		fmt.Println("File Chunk", noChunks, subdomain, time.Since(startTime))
		time.Sleep(10 * time.Millisecond)
	}

	indexFile := fmt.Sprintf("%s.%s", createIndexRecord.Subdomain, f.config.Domain)
	fmt.Println("Total Time Taken", time.Since(startTime))
	return indexFile, nil
}

func (f *FileUploader) Download(ctx context.Context, indexFileRecord string) (string, error) {
	startTime := time.Now()
	// Get TXT Record
	txtRecord, err := f.dnsClient.ReadTXTRecord(indexFileRecord)
	if err != nil {
		return "", err
	}

	subdomain := strings.Split(indexFileRecord, "."+f.config.Domain)[0]

	// TXT Records: <NAME.FILE_TYPE>.<END_CHUNK|100>
	temp := strings.Split(txtRecord, ".")
	fileName := strings.Join(temp[:len(temp)-1], ".")
	noChunks, err := strconv.Atoi(temp[len(temp)-1])
	if err != nil {
		return "", err
	}
	fmt.Println("FileName:", fileName)
	fmt.Println("Total noChunks:", noChunks)
	fmt.Println("Total Time Taken", time.Since(startTime))

	downloadPath := filepath.Join(f.config.DownloadDir, fileName)
	var file *os.File
	file, err = os.OpenFile(downloadPath, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	fmt.Println("FileStats", fileInfo.Size())
	downloadChunks := int(math.Ceil(float64(fileInfo.Size()) / float64(MaxChunkSize)))
	fmt.Println("DownloadChunks", downloadChunks)
	for i := downloadChunks; i < noChunks; {
		domain := fmt.Sprintf("%s.%d.%s", subdomain, i, f.config.Domain)
		txtRecord, err := f.dnsClient.ReadTXTRecord(domain)
		if err != nil {
			fmt.Println("Retrying: Error Reading", err)
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		rawBinary, err := base64.StdEncoding.DecodeString(txtRecord)
		if err != nil {
			return "", err
		}
		n, err := file.WriteAt(rawBinary, int64(i)*MaxChunkSize)
		if err != nil {
			return "", err
		}
		if err := file.Sync(); err != nil {
			return "", err
		}
		fmt.Println("File Chunk", i, " bytes written ", n, time.Since(startTime))
		time.Sleep(25 * time.Millisecond)
		i++
	}

	fmt.Println("Total Time Taken", time.Since(startTime))
	return downloadPath, nil
}

func (f *FileUploader) Delete(ctx context.Context, indexFileRecord string) error {
	// Check correct index file record
	// Get TXT Record
	txtRecord, err := f.dnsClient.ReadTXTRecord(indexFileRecord)
	if err != nil {
		return err
	}

	subdomain := strings.Split(indexFileRecord, "."+f.config.Domain)[0]

	fmt.Println("TXT Record:", txtRecord)
	temp := strings.Split(txtRecord, ".")
	noChunks, err := strconv.Atoi(temp[len(temp)-1])
	if err != nil {
		return err
	}

	for i := 0; i < noChunks; {
		subdomain := fmt.Sprintf("%s.%d", subdomain, i)
		record, err := f.dnsProviderClient.GetTXTRecords(ctx, subdomain)
		if err != nil {
			fmt.Println("Retrying: Error Reading", err)
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		err = f.dnsProviderClient.DeleteTXTRecord(ctx, record.ID)
		if err != nil {
			fmt.Println("Retrying: Error Deleting", err)
			time.Sleep(1000 * time.Millisecond)
			continue
		}
		fmt.Println("Deleted", subdomain)
		time.Sleep(25 * time.Millisecond)
		i++
	}

	indexRecord, err := f.dnsProviderClient.GetTXTRecords(ctx, subdomain)
	if err != nil {
		return err
	}
	err = f.dnsProviderClient.DeleteTXTRecord(ctx, indexRecord.ID)
	if err != nil {
		return err
	}
	fmt.Println("Deleted index File", indexFileRecord)
	return nil
}

func (f *FileUploader) Stream(ctx context.Context, record string) error {
	// I do not know it
	panic("implement me")
	return nil
}
