// provides RSA open/closed key pair encryption
package cryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
)

// generates a new RSA key pair of configurable length and saves them to files
func GenerateKeyPair(privateKeyFile, publicKeyFile string, bits int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	// Save private key
	privFile, err := os.Create(privateKeyFile)
	if err != nil {
		return err
	}
	defer privFile.Close()

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	pem.Encode(privFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	// Save public key
	pubFile, err := os.Create(publicKeyFile)
	if err != nil {
		return err
	}
	defer pubFile.Close()

	publicKey := &privateKey.PublicKey
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}

	return pem.Encode(pubFile, &pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes})
}

// loads private key from file provided
func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// loads public key from file provided
func ReadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyData)

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key.(*rsa.PublicKey), nil
}

func EncryptRaw(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
}

func DecryptRaw(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
}

func EncryptStr(data []byte, publicKey *rsa.PublicKey) (string, error) {
	encryptedBytes, err := EncryptRaw(data, publicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

func DecryptStr(encryptedData string, privateKey *rsa.PrivateKey) ([]byte, error) {
	decodedData, _ := base64.StdEncoding.DecodeString(encryptedData)
	return DecryptRaw(decodedData, privateKey)
}
