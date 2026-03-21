package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"dns-storage/pkg/defaults"
)

const (
	defaultMaxMemory = 4 << 20 // 4 MB
)

type HTTPHandler struct {
	config            *defaults.DefaultConfig
	dnsProviderClient DNSTXTProvider
	dnsClient         DNSTXTHandler
	fileHandler       FileHandlerProvider
}

func NewHTTPHandler(
	config *defaults.DefaultConfig,
	dnsProviderCli DNSTXTProvider,
	dnsCli DNSTXTHandler,
	fileHandler FileHandlerProvider,
) APIHandler {
	return &HTTPHandler{
		config:            config,
		dnsProviderClient: dnsProviderCli,
		dnsClient:         dnsCli,
		fileHandler:       fileHandler,
	}
}

// Health implements [APIHandler].
func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	health := `{
	"health": "Server is running"
}`
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(health))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Download implements [APIHandler].
func (h *HTTPHandler) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	indexFileRecord := r.URL.Query().Get("file")
	fmt.Println("indexFileRecord:", indexFileRecord)

	isStream := r.URL.Query().Get("stream")
	if isStream != "true" {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	fileChan, errChan := h.fileHandler.Stream(r.Context(), indexFileRecord)

	isFileTypeHeaderPresent := false
	for {
		var err error

		select {
		case stream, ok := <-fileChan:
			if !ok {
				return
			}
			if !isFileTypeHeaderPresent && isStream != "true" {
				w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, stream.MetaData.FileName))
				isFileTypeHeaderPresent = true
			} else if !isFileTypeHeaderPresent {
				fileTypeHeader := getFileTypeHeaderFromName(stream.MetaData.FileName)
				w.Header().Set("Content-Type", fileTypeHeader)
				isFileTypeHeaderPresent = true
			}

			_, err = w.Write(stream.Data)
			if err != nil {
				break
			}
			fmt.Println("status:", stream.MetaData.CurrentChunk, " / ", stream.MetaData.TotalChunks)
			fmt.Println("subdomain:", stream.MetaData.Subdomain)
			fmt.Println("fileName:", stream.MetaData.FileName)
		case err1, ok := <-errChan:
			if !ok {
				break
			}
			err = err1
		case <-r.Context().Done():
			err = r.Context().Err()
		}

		if err != nil {
			fmt.Println("err:", err)
			// http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Upload implements [APIHandler].
func (h *HTTPHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileStatus, err := h.processUpload(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	response := fmt.Sprintf(`{
	"status": "success",
	"indexRecord": "%s.%s"
	}`, fileStatus.Subdomain, h.config.Domain)

	_, err = w.Write([]byte(response))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *HTTPHandler) processUpload(r *http.Request) (FileStatus, error) {
	// TODO: Implement MultipartReader directly process the byte and create TXT reacord
	// reader, err := r.MultipartReader()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
	//
	startTime := time.Now()
	var fileStatus FileStatus

	if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
		return fileStatus, err
	}
	multipartForm := r.MultipartForm
	fmt.Printf("\n%#v\n", multipartForm)

	values := multipartForm.Value
	subdomains, ok := values["Subdomain"]
	if !ok {
		return fileStatus, errors.New("subdomain not found")
	}
	if len(subdomains) > 1 {
		return fileStatus, errors.New("more than one subdomain found")
	}
	subdomain := subdomains[0]
	fmt.Println("Subdomain:", subdomain)

	if len(multipartForm.File) > 1 {
		return fileStatus, errors.New("more than one file found")
	}

	files, ok := multipartForm.File["File"]
	if !ok {
		return fileStatus, errors.New("file not found")
	}

	if len(files) > 1 {
		return fileStatus, errors.New("more than one file found")
	}

	fileInfo := files[0]
	fmt.Println("File:", fileInfo.Filename)
	fmt.Println("Size:", fileInfo.Size)

	file, err := fileInfo.Open()
	if err != nil {
		return fileStatus, err
	}
	defer file.Close()

	ctx := r.Context()
	name := fileInfo.Filename
	totalChunks := int(math.Ceil(float64(fileInfo.Size) / float64(h.config.GetMaxChunkByteSize())))
	fmt.Println("Total Chunks:", totalChunks)

	fileStatus = FileStatus{
		TotalChunks:  totalChunks,
		CurrentChunk: 0,
		Subdomain:    subdomain,
	}

	// Create a index TXT record for that chuck file
	// TXT Records: <NAME.FILE_TYPE>.<END_CHUNK_NO>
	txtRecord := fmt.Sprintf("%s.%d", name, fileStatus.TotalChunks)
	createIndexRecord, err := h.dnsProviderClient.CreateTXTRecord(ctx, subdomain, txtRecord)
	if err != nil {
		return fileStatus, err
	}
	// If some error occur update the index file txt record properly
	defer func() {
		txtRecord := fmt.Sprintf("%s.%d", name, fileStatus.CurrentChunk)
		newRecord := createIndexRecord
		newRecord.Content = txtRecord
		_, err := h.dnsProviderClient.UpdateTXTRecord(ctx, strconv.Itoa(newRecord.ID), newRecord)
		if err != nil {
			fmt.Println("Error while creating last chunk", err)
		}
		fmt.Println("File Last chunk Complete", createIndexRecord)
	}()

	data := make([]byte, h.config.GetMaxChunkByteSize())
	var offset int64 = 0

	for {
		n, err := file.ReadAt(data, offset)
		if n == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return fileStatus, err
		}
		data = data[:n]
		subdomain := fmt.Sprintf("%s.%d", subdomain, fileStatus.CurrentChunk)
		base64Value := base64.StdEncoding.EncodeToString(data)
		_, err = h.dnsProviderClient.CreateTXTRecord(ctx, subdomain, base64Value)
		if err != nil {
			return fileStatus, err
		}
		offset += int64(n)

		fileStatus.Subdomain = subdomain
		fmt.Println("File Chunk", fileStatus.CurrentChunk, " bytes written ", n, time.Since(startTime))
		fileStatus.CurrentChunk++
	}

	indexFile := fmt.Sprintf("%s.%s", createIndexRecord.Subdomain, h.config.Domain)
	fmt.Println("Total Time Taken", time.Since(startTime))
	fmt.Println("Index File domain", indexFile)
	return fileStatus, nil
}
