package utils

import (
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"io"
	"net/http"
)

func GetBodyBytes(r *http.Request) ([]byte, error) {
	var reader io.Reader

	if r.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			return []byte{}, err
		}
		defer gzipReader.Close()
		reader = gzipReader
	} else if r.Header.Get("Content-Encoding") == "br" {
		brReader := brotli.NewReader(r.Body)
		reader = brReader
	} else {
		reader = r.Body
	}
	defer r.Body.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}
