package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"dns-storage/pkg/defaults"

	"github.com/google/uuid"
)

func NewFileHander(
	config *defaults.DefaultConfig,
	dnsProvierCli DNSTXTProvider,
	dnsCli DNSTXTHandler,
) FileHandlerProvider {
	return &FileUploader{
		config:            config,
		dnsProviderClient: dnsProvierCli,
		dnsClient:         dnsCli,
	}
}

type FileUploader struct {
	config            *defaults.DefaultConfig
	dnsProviderClient DNSTXTProvider
	dnsClient         DNSTXTHandler
}

// Delete implements [FileHandlerProvider].
func (f *FileUploader) Delete(ctx context.Context, indexFileRecord string) (<-chan FileStatus, <-chan error) {
	fileStatusChan, errChan := make(chan FileStatus), make(chan error)

	go func() {
		defer close(fileStatusChan)
		defer close(errChan)
		// Check correct index file record
		// Get TXT Record
		txtRecord, err := f.dnsClient.ReadTXTRecord(ctx, indexFileRecord)
		if err != nil {
			errChan <- err
			return
		}

		subdomain := strings.Split(indexFileRecord, "."+f.config.Domain)[0]

		fmt.Println("TXT Record:", txtRecord)
		temp := strings.Split(txtRecord, ".")
		fileName := strings.Join(temp[:len(temp)-1], ".")
		totalChunks, err := strconv.Atoi(temp[len(temp)-1]) // 0 based index
		if err != nil {
			errChan <- err
			return
		}

		fileStatus := FileStatus{
			TotalChunks:  totalChunks,
			CurrentChunk: 0,
			Subdomain:    subdomain,
			FileName:     fileName,
		}

		for fileStatus.CurrentChunk < fileStatus.TotalChunks {
			subdomain := fmt.Sprintf("%s.%d", subdomain, fileStatus.CurrentChunk)
			record, err := f.dnsProviderClient.GetTXTRecords(ctx, subdomain)
			if err != nil {
				fmt.Println("Retrying: Error Reading", err, subdomain)
				errChan <- err
				return
			}
			err = f.dnsProviderClient.DeleteTXTRecord(ctx, strconv.Itoa(record.ID))
			if err != nil {
				fmt.Println("Retrying: Error Deleting", err)
				errChan <- err
				return
			}
			fmt.Println("Deleted", subdomain)

			fileStatus.Subdomain = subdomain
			fileStatus.CurrentChunk++
			fileStatusChan <- fileStatus
		}

		indexRecord, err := f.dnsProviderClient.GetTXTRecords(ctx, subdomain)
		if err != nil {
			errChan <- err
			return
		}
		err = f.dnsProviderClient.DeleteTXTRecord(ctx, strconv.Itoa(indexRecord.ID))
		if err != nil {
			errChan <- err
			return
		}
		fileStatus.Subdomain = subdomain
		fileStatusChan <- fileStatus
		fmt.Println("Deleted index File", indexFileRecord)
	}()

	return fileStatusChan, errChan
}

// Download implements [FileHandlerProvider].
func (f *FileUploader) Download(ctx context.Context, indexFileRecord string, filePath string) (<-chan FileStatus, <-chan error) {
	fileStatusChan, errChan := make(chan FileStatus), make(chan error)

	go func() {
		defer close(fileStatusChan)
		defer close(errChan)

		startTime := time.Now()
		// Get TXT Record
		txtRecord, err := f.dnsClient.ReadTXTRecord(ctx, indexFileRecord)
		if err != nil {
			fmt.Println("ReadTXTRecord Error for :", indexFileRecord)
			errChan <- err
			return
		}

		subdomain := strings.Split(indexFileRecord, "."+f.config.Domain)[0]

		// TXT Records: <NAME.FILE_TYPE>.<END_CHUNK|100>
		temp := strings.Split(txtRecord, ".")
		fileName := strings.Join(temp[:len(temp)-1], ".")
		totalChunks, err := strconv.Atoi(temp[len(temp)-1]) // 0 based index
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("FileName:", fileName)
		fmt.Println("Total noChunks:", totalChunks)
		fmt.Println("Total Time Taken", time.Since(startTime))

		fileStatus := FileStatus{
			TotalChunks:  totalChunks,
			CurrentChunk: 0,
			Subdomain:    subdomain,
			FileName:     fileName,
		}

		var file *os.File
		file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o666)
		if err != nil {
			errChan <- err
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			errChan <- err
			return
		}

		fmt.Println("FileStats", fileInfo.Size())
		downloadChunks := int(math.Ceil(float64(fileInfo.Size()) / float64(f.config.GetMaxChunkByteSize())))
		fmt.Println("DownloadChunks", downloadChunks)

		fileStatus.CurrentChunk = downloadChunks

		retryCounter := 0
		for fileStatus.CurrentChunk < fileStatus.TotalChunks {
			domain := fmt.Sprintf("%s.%d.%s", subdomain, fileStatus.CurrentChunk, f.config.Domain)
			txtRecord, err := f.dnsClient.ReadTXTRecord(ctx, domain)
			if err != nil {
				fmt.Println("Retrying: Error Reading", err, domain, retryCounter)
				if retryCounter < f.config.DNSRetryLimit {
					retryCounter++
					continue
				}
				errChan <- err
				return
			}
			retryCounter = 0

			rawBinary, err := base64.StdEncoding.DecodeString(txtRecord)
			if err != nil {
				errChan <- err
				return
			}
			n, err := file.WriteAt(rawBinary, int64(fileStatus.CurrentChunk)*int64(f.config.GetMaxChunkByteSize()))
			if err != nil {
				errChan <- err
				return
			}
			if err := file.Sync(); err != nil {
				errChan <- err
				return
			}

			fileStatus.Subdomain = fmt.Sprintf("%s.%d", subdomain, fileStatus.CurrentChunk)
			fileStatus.CurrentChunk++
			fileStatusChan <- fileStatus

			fmt.Println("File Chunk", fileStatus.CurrentChunk, " bytes written ", n, time.Since(startTime))
		}

		fmt.Println("Total Time Taken", time.Since(startTime))
	}()

	return fileStatusChan, errChan
}

// Stream implements [FileHandlerProvider].
func (f *FileUploader) Stream(ctx context.Context, indexFileRecord string) (<-chan FileStream, <-chan error) {
	fileStatusChan, errChan := make(chan FileStream), make(chan error)

	go func() {
		defer close(fileStatusChan)
		defer close(errChan)

		startTime := time.Now()
		// Get TXT Record
		txtRecord, err := f.dnsClient.ReadTXTRecord(ctx, indexFileRecord)
		if err != nil {
			errChan <- err
			return
		}

		subdomain := strings.Split(indexFileRecord, "."+f.config.Domain)[0]

		// TXT Records: <NAME.FILE_TYPE>.<END_CHUNK|100>
		temp := strings.Split(txtRecord, ".")
		fileName := strings.Join(temp[:len(temp)-1], ".")
		totalChunks, err := strconv.Atoi(temp[len(temp)-1]) // 0 based index
		if err != nil {
			errChan <- err
			return
		}
		fmt.Println("FileName:", fileName)
		fmt.Println("Total noChunks:", totalChunks)
		fmt.Println("Total Time Taken", time.Since(startTime))

		fileStream := FileStream{
			Data: make([]byte, 0),
			MetaData: FileStatus{
				TotalChunks:  totalChunks,
				CurrentChunk: 0,
				Subdomain:    subdomain,
				FileName:     fileName,
				BatchSize:    f.config.StreamBatchSize,
			},
		}

		for fileStream.MetaData.CurrentChunk < fileStream.MetaData.TotalChunks {
			streamDataInOrder := make([][]byte, fileStream.MetaData.BatchSize)

			isError := false
			wg := sync.WaitGroup{}
			for i := 0; i < fileStream.MetaData.BatchSize && i+fileStream.MetaData.CurrentChunk < fileStream.MetaData.TotalChunks; i++ {
				wg.Add(1)
				go func(j, currentChunk int, dataStore [][]byte) {
					retryCounter := 0
					defer wg.Done()

					var txtRecord string
					var err error
					for {
						domain := fmt.Sprintf("%s.%d.%s", subdomain, j+currentChunk, f.config.Domain)
						txtRecord, err = f.dnsClient.ReadTXTRecord(ctx, domain)
						if err != nil {
							fmt.Println("Retrying: Error Reading", err, domain, retryCounter)
							if retryCounter < f.config.DNSRetryLimit {
								retryCounter++
								continue
							}
							errChan <- err
							isError = true
							return
						}
						break
					}

					rawBinary, err := base64.StdEncoding.DecodeString(txtRecord)
					if err != nil {
						isError = true
						errChan <- err
						return
					}
					dataStore[j] = rawBinary
				}(i, fileStream.MetaData.CurrentChunk, streamDataInOrder)
			}
			wg.Wait()
			if isError {
				fmt.Println("Error while reading data in the batch")
				return
			}

			rawBinary := []byte{}
			for _, data := range streamDataInOrder {
				rawBinary = append(rawBinary, data...)
			}

			fileStream.Data = rawBinary
			fileStream.MetaData.Subdomain = fmt.Sprintf("%s.%d", subdomain, fileStream.MetaData.CurrentChunk)
			if fileStream.MetaData.CurrentChunk+fileStream.MetaData.BatchSize > fileStream.MetaData.TotalChunks {
				fileStream.MetaData.CurrentChunk = fileStream.MetaData.TotalChunks - fileStream.MetaData.BatchSize
			}
			fileStream.MetaData.CurrentChunk += fileStream.MetaData.BatchSize
			fileStatusChan <- fileStream
		}

		fmt.Println("Total Time Taken", time.Since(startTime))
	}()

	return fileStatusChan, errChan
}

// Upload implements [FileHandlerProvider].
func (f *FileUploader) Upload(ctx context.Context, filePath string, subdomain string) (<-chan FileStatus, <-chan error) {
	fileStatusChan, errChan := make(chan FileStatus), make(chan error)

	go func() {
		defer close(fileStatusChan)
		defer close(errChan)

		// Calculate the no of chunks need to create
		fmt.Println("FilePath:", filePath)
		file, err := os.Open(filePath)
		if err != nil {
			errChan <- err
			return
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			errChan <- err
			return
		}

		startTime := time.Now()
		if subdomain == "" {
			subdomain = uuid.New().String()
		}
		name := fileInfo.Name()
		totalChunks := int(math.Ceil(float64(fileInfo.Size()) / float64(f.config.GetMaxChunkByteSize())))
		fmt.Println("Total Chunks:", totalChunks)

		fileStatus := FileStatus{
			TotalChunks:  totalChunks,
			CurrentChunk: 0,
			Subdomain:    subdomain,
			FileName:     fileInfo.Name(),
		}

		// Create a index TXT record for that chuck file
		// TXT Records: <NAME.FILE_TYPE>.<END_CHUNK_NO>
		txtRecord := fmt.Sprintf("%s.%d", name, totalChunks)
		createIndexRecord, err := f.dnsProviderClient.CreateTXTRecord(ctx, subdomain, txtRecord)
		if err != nil {
			errChan <- err
			return
		}
		//
		data := make([]byte, f.config.GetMaxChunkByteSize())
		noChunks := 0
		var offset int64 = 0

		// If some error occur update the index file txt record properly
		defer func() {
			txtRecord := fmt.Sprintf("%s.%d", name, noChunks)
			newRecord := createIndexRecord
			newRecord.Content = txtRecord
			_, err := f.dnsProviderClient.UpdateTXTRecord(ctx, strconv.Itoa(newRecord.ID), newRecord)
			if err != nil {
				fmt.Println("Error while creating last chunk", err)
			}
			fmt.Println("File Last chunk Complete", createIndexRecord)
		}()

		for {
			n, err := file.ReadAt(data, offset)
			if n == 0 && err != nil {
				if err == io.EOF {
					break
				}
				errChan <- err
				return
			}
			data = data[:n]
			subdomain := fmt.Sprintf("%s.%d", subdomain, fileStatus.CurrentChunk)
			base64Value := base64.StdEncoding.EncodeToString(data)
			_, err = f.dnsProviderClient.CreateTXTRecord(ctx, subdomain, base64Value)
			if err != nil {
				errChan <- err
				return
			}
			offset += int64(n)

			fileStatus.Subdomain = subdomain
			fileStatusChan <- fileStatus

			fmt.Println("File Chunk", noChunks, subdomain, time.Since(startTime))
			fileStatus.CurrentChunk++
		}

		indexFile := fmt.Sprintf("%s.%s", createIndexRecord.Subdomain, f.config.Domain)
		fmt.Println("Total Time Taken", time.Since(startTime))
		fmt.Println("Index File domain", indexFile)
	}()

	return fileStatusChan, errChan
}
