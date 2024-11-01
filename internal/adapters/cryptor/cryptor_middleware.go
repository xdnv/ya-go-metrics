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
func HandleencryptedRequests(next http.Handler) http.Handler {
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

		//passing body to next handler
		body, _ := io.ReadAll(r.Body)
		r.Body.Close() //  must close
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		if len(body) == 0 {
			logger.Info("cryptor: empty body, no decryption required")
			next.ServeHTTP(rw, r)
			return
		}

		logger.Info("cryptor: handling encrypted request")

		decrBody, err := Decrypt(body)
		if err != nil {
			logger.Error("cryptor: error decrypting payload: " + err.Error())
			http.Error(rw, "server could not read encrypted message", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(decrBody))

		logger.Info("cryptor: successfully decrypted")

		next.ServeHTTP(rw, r)
	})
}
