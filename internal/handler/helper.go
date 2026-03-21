package handler

import (
	"strings"
)

func getFileTypeHeaderFromName(fileName string) string {
	temp := strings.Split(fileName, ".")
	fileType := temp[len(temp)-1]
	fileTypeHeader := "text/plain"
	switch fileType {
	// Text formats
	// case "txt":
	// 	fileTypeHeader = "text/plain"
	case "html":
		fileTypeHeader = "text/html"
	case "json":
		fileTypeHeader = "application/json"

	// Image formats
	case "png":
		fileTypeHeader = "image/png"
	case "jpg":
		fileTypeHeader = "image/jpeg"
	case "jpeg":
		fileTypeHeader = "image/jpeg"
	case "gif":
		fileTypeHeader = "image/gif"
	case "webp":
		fileTypeHeader = "image/webp"
	case "svg":
		fileTypeHeader = "image/svg+xml"
	case "avif":
		fileTypeHeader = "image/avif"

	// PDF format
	case "pdf":
		fileTypeHeader = "application/pdf"

	// Video formats
	case "mp4":
		fileTypeHeader = "video/mp4"
	case "webm":
		fileTypeHeader = "video/mpeg"
	case "ogg":
		fileTypeHeader = "video/ogg"

	// Audio formats
	case "mpeg":
		fileTypeHeader = "audio/mpeg"
	case "wav":
		fileTypeHeader = "audio/wav"
	case "mp3":
		fileTypeHeader = "audio/mpeg"
	}
	return fileTypeHeader
}
