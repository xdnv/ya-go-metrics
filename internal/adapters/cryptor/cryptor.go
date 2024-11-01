package cryptor

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
)

// main cryptor object to store security configuration
type CryptorObject struct {
	UseEncryption bool            // enables or disables encryption (regardless of keys loaded)
	PrivateKey    *rsa.PrivateKey // secret key to en/decode payload
	PublicKey     *rsa.PublicKey  // public key to encode payload
}

type PayloadObject struct {
	Key     string `json:"key"`
	Payload string `json:"payload"`
}

var cryptor *CryptorObject

func init() {
	cryptor = new(CryptorObject)
}

func EnableEncryption(encrypt bool) {
	cryptor.UseEncryption = encrypt
}

func CanEncrypt() bool {
	return cryptor.UseEncryption && cryptor.PublicKey != nil
}

func CanDecrypt() bool {
	return cryptor.UseEncryption && cryptor.PrivateKey != nil
}

// loads private & public keys from file provided
func LoadPrivateKey(path string) error {
	keyData, err := ReadPrivateKey(path)
	if err != nil {
		return err
	}

	cryptor.PrivateKey = keyData
	cryptor.PublicKey = &keyData.PublicKey

	return nil
}

// loads public key from file provided
func LoadPublicKey(path string) error {
	keyData, err := ReadPublicKey(path)
	if err != nil {
		return err
	}

	cryptor.PublicKey = keyData

	return nil
}

func Encrypt(data []byte) ([]byte, error) {
	// Generate random 256-bit AES key
	aesKey, err := GenerateAESKey(32)
	if err != nil {
		return nil, err
	}

	// Encrypt message using AES
	encryptedMessage, err := EncryptAES(aesKey, data)
	if err != nil {
		return nil, err
	}

	// Encrypt AES key using RSA
	encryptedAESKey, err := EncryptRaw(aesKey, cryptor.PublicKey)
	if err != nil {
		return nil, err
	}

	//build payload data structure
	payload := PayloadObject{
		Key:     base64.StdEncoding.EncodeToString(encryptedAESKey),
		Payload: encryptedMessage,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func Decrypt(encryptedData []byte) ([]byte, error) {
	var payload PayloadObject

	err := json.Unmarshal(encryptedData, &payload)
	if err != nil {
		return nil, err
	}

	// Decode base64 encoded AES key
	encrAESKey, err := base64.StdEncoding.DecodeString(payload.Key)
	if err != nil {
		return nil, err
	}

	aesKey, err := DecryptRaw(encrAESKey, cryptor.PrivateKey)
	if err != nil {
		return nil, err
	}

	decryptedMessage, err := DecryptAES(aesKey, payload.Payload)
	if err != nil {
		return nil, err
	}

	return decryptedMessage, nil
}

// get HTTP request header used to get encryption mark
func GetEncryptionToken() string {
	return "X-Encrypted"
}
