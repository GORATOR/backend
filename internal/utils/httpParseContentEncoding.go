package utils

import (
	"compress/gzip"
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
