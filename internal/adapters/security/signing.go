package security

import (
	"crypto/hmac"
	"crypto/sha256"
)

func GetSignature(payload []byte, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(payload)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func GetSignatureToken() string {
	return "HashSHA256"
}
