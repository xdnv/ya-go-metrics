package signer

import (
	"crypto/hmac"
	"crypto/sha256"
)

type SignerObject struct {
	UseSignedMessaging    bool
	StrictSignedMessaging bool
	MsgKey                string
}

var signer *SignerObject

func init() {
	signer = new(SignerObject)
	signer.StrictSignedMessaging = false
}

func SetKey(msgKey string) {
	signer.MsgKey = msgKey
	signer.UseSignedMessaging = (signer.MsgKey != "")
}

func UseSignedMessaging() bool {
	return signer.UseSignedMessaging
}

func GetSignature(payload []byte) ([]byte, error) {
	key := []byte(signer.MsgKey)
	return GetSignatureByKey(payload, key)
}

func GetSignatureByKey(payload []byte, key []byte) ([]byte, error) {
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
