package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"internal/adapters/logger"
	"io"
	"net/http"
	"strings"
)

func handleGZIPRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(rw, r)
			return
		}

		logger.Info("srv-gzip: handling gzipped request")

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			errText := fmt.Sprintf("error reading compressed msg body: %s", err.Error())
			logger.Error("handleGZIPRequests: " + errText)
			http.Error(rw, errText, http.StatusInternalServerError)
			return
		}

		defer gz.Close()
		body, err := io.ReadAll(gz)
		if err != nil {
			errText := fmt.Sprintf("error extracting msg body: %s", err.Error())
			logger.Error("handleGZIPRequests: " + errText)
			http.Error(rw, errText, http.StatusInternalServerError)
			return
		}

		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(rw, r)
	})
}
