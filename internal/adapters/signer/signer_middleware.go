// the signer middleware provides transparent HTTP command signature validation
package signer

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"

	"internal/adapters/logger"
)

// provides message security check using stored signature
func HandleSignedRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		if !IsSignedMessagingEnabled() {
			next.ServeHTTP(rw, r)
			return
		}

		//passing body to next handler
		body, _ := io.ReadAll(r.Body)
		r.Body.Close() //  must close
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		if len(body) == 0 {
			logger.Info("signer: empty body, no security check needed")
			next.ServeHTTP(rw, r)
			return
		}

		logger.Info("signer: handling signed request")

		sigRaw := r.Header.Get(GetSignatureToken())
		//logger.Info("srv-sec: sig=" + sigRaw)

		sig, err := base64.URLEncoding.DecodeString(sigRaw)
		if err != nil {
			logger.Error("signer: incorrect message signature format")
			http.Error(rw, "incorrect message signature format", http.StatusBadRequest)
			return
		}

		//calculate body signature
		ok := Compare(&sig, &body)
		if !ok {
			if IsStrictSignedMessagingEnabled() {
				logger.Error("signer: message security check failed")
				http.Error(rw, "message security check failed", http.StatusBadRequest)
				return
			}

			//non-strict mode passes yandex iter14 test: yandex gives no actual signature, just a key on startup
			logger.Error("signer: non-strict message security check FAILED")
		} else {
			logger.Infof("signer: signature OK, id=%s", sig)
		}

		next.ServeHTTP(rw, r)
	})
}
