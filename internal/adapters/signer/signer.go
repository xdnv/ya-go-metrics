// the signer module provides sha256 signature check for any binary payload
package signer

import (
	"crypto/hmac"
	"crypto/sha256"
)

// main signer object to store security configuration
type SignerObject struct {
	UseSignedMessaging    bool   // enables or disables use of signature
	StrictSignedMessaging bool   // YP compatibility flag to pass failed check with warning
	MsgKey                string // secret key to encode payload
}

var signer *SignerObject

func init() {
	signer = new(SignerObject)
	signer.StrictSignedMessaging = true
}

// set security key
func SetKey(msgKey string) {
	signer.MsgKey = msgKey
	signer.UseSignedMessaging = (signer.MsgKey != "")
}

// return security state of the signer module
func UseSignedMessaging() bool {
	return signer.UseSignedMessaging
}

// get signature for binary payload using stored security key
func GetSignature(payload []byte) ([]byte, error) {
	key := []byte(signer.MsgKey)
	return GetSignatureByKey(payload, key)
}

// get signature for binary payload using provided security key
func GetSignatureByKey(payload []byte, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(payload)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// get HTTP request header used to store signature
func GetSignatureToken() string {
	return "HashSHA256"
}
