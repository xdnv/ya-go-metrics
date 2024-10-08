package signer_test

import (
	"crypto/hmac"
	"fmt"
	"signer"
)

// HMAC may be used to protect data from tampering attacks
func ExampleGetSignature() {

	payload := "Hash me!"
	key := "secret"

	sig1, err := signer.GetSignatureByKey([]byte(payload), []byte(key))
	if err != nil {
		fmt.Println(err)
		return
	}

	sig2, err := signer.GetSignatureByKey([]byte(payload), []byte(key))
	if err != nil {
		fmt.Println(err)
		return
	}

	sig3, err := signer.GetSignatureByKey([]byte(payload+"?"), []byte(key))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("equal with original:", hmac.Equal(sig1, sig2))
	fmt.Println("equal with tampered:", hmac.Equal(sig1, sig3))

	// Output:
	// equal with original: true
	// equal with tampered: false
}
