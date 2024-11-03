// the cryptor middleware provides transparent HTTP command decryption (RSA+AES)
package cryptor

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"internal/adapters/logger"
)

// provides message security check using stored signature
func HandleEncryptedRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		// check for corresponding encryption header
		encRaw := strings.TrimSpace(r.Header.Get(GetEncryptionToken()))
		if encRaw != "true" {
			next.ServeHTTP(rw, r)
			return
		}

		if !CanDecrypt() {
			logger.Error("cryptor: server is not configured to read encrypted messages")
			http.Error(rw, "server is not configured to read encrypted messages", http.StatusBadRequest)
			return
		}

		//reading out body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("cryptor: error reading encrypted request body: " + err.Error())
			http.Error(rw, "error reading encrypted request body", http.StatusInternalServerError)
			return
		}
		r.Body.Close() //  must close

		if len(body) == 0 {
			logger.Info("cryptor: empty body, no decryption required")

			//passing body to next handler
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			//r.Body = io.NopCloser(io.Reader(bytes.NewReader(body)))
			next.ServeHTTP(rw, r)
			return
		}

		logger.Info("cryptor: handling encrypted request")

		decrBody, err := Decrypt(&body)
		if err != nil {
			logger.Error("cryptor: error decrypting payload: " + err.Error())
			http.Error(rw, "server could not read encrypted message", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(decrBody))
		//r.Body = io.NopCloser(io.Reader(bytes.NewReader(decrBody)))

		logger.Info("cryptor: successfully decrypted")

		next.ServeHTTP(rw, r)
	})
}
