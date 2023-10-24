package util

import (
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/h2non/filetype"
)

func GetFileExt(fileData []byte) string {
	kind, _ := filetype.Match(fileData)
	if kind != filetype.Unknown && kind.Extension != "" {
		return kind.Extension
	}
	//fallback to DetectContentType
	mimeType := http.DetectContentType(fileData)
	return mimeType[strings.LastIndex(mimeType, "/")+1:]
}

func WriteFile(writer *multipart.Writer, fileName string, fileData []byte) error {
	part, err := writer.CreateFormFile(fileName, "file."+GetFileExt(fileData))
	if err != nil {
		return err
	}
	_, err = part.Write(fileData)
	return err
}

func WriteField(writer *multipart.Writer, key string, value string) {
	if key != "" && value != "" {
		_ = writer.WriteField(key, value)
	}
}
