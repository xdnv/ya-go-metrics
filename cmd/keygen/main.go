package main

import (
	"fmt"
	"os"

	"internal/adapters/cryptor"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: keygen <private_key_file> <public_key_file>")
		return
	}

	privateKeyFile := os.Args[1]
	publicKeyFile := os.Args[2]

	if err := cryptor.GenerateKeyPair(privateKeyFile, publicKeyFile, 4096); err != nil {
		fmt.Println("Error generating keys:", err)
		return
	}
	fmt.Println("RSA keys generated and saved to", privateKeyFile+" and "+publicKeyFile)
}
