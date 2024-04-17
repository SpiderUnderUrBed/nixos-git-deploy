package aged

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"bytes"

	"filippo.io/age"
	"filippo.io/age/armor"
)

func Decrypt() {
	identityBytes, err := os.ReadFile("privatekey.txt")
	if err != nil {
		log.Fatalf("Failed to read identity file: %v", err)
	}
	receiver, err := age.ParseX25519Identity(string(identityBytes))
	if err != nil {
		log.Fatalf("Failed to parse X25519 identity: %v", err)
	}

	// Read the encrypted file and decrypt the message
	encryptedBytes, err := os.ReadFile("encrypted.txt")
	if err != nil {
		log.Fatalf("Failed to read encrypted file: %v", err)
	}
	armorReader := armor.NewReader(bytes.NewReader(encryptedBytes))
	r, err := age.Decrypt(armorReader, receiver)
	if err != nil {
		log.Fatalf("Failed to decrypt message: %v", err)
	}

	// Write the decrypted message to stdout
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatalf("Failed to write decrypted message: %v", err)
	}
	fmt.Println()
}

func Encrypt() {
	msg := "Hello"

	// Check if the private key file exists
	privateKeyFile := "privatekey.txt"
	privateKey, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		fmt.Println("Not using keyfile")
		if !os.IsNotExist(err) {
			log.Fatalf("Failed to read private key file: %v", err)
		}

		// Generate X25519 identity if private key file does not exist
		identity, err := age.GenerateX25519Identity()
		if err != nil {
			log.Fatalf("Failed to generate X25519 identity: %v", err)
		}
		privateKey = []byte(identity.String())

		// Save the private key to a file
		if err := ioutil.WriteFile(privateKeyFile, privateKey, 0644); err != nil {
			log.Fatalf("Failed to write private key to file: %v", err)
		}
	} else {
		fmt.Println("Using keyfile")
	}

	// Check if the public key file exists
	publicKeyFile := "publickey.txt"
	publicKey, err := ioutil.ReadFile(publicKeyFile)
	if err != nil {
		fmt.Println("Not using keyfile")
		if !os.IsNotExist(err) {
			log.Fatalf("Failed to read public key file: %v", err)
		}

		// Generate X25519 identity if public key file does not exist
		identity, err := age.GenerateX25519Identity()
		if err != nil {
			log.Fatalf("Failed to generate X25519 identity: %v", err)
		}
		publicKey = []byte(identity.Recipient().String())

		// Save the public key to a file
		if err := ioutil.WriteFile(publicKeyFile, publicKey, 0644); err != nil {
			log.Fatalf("Failed to write public key to file: %v", err)
		}
	} else {
		fmt.Println("Using keyfile")
	}

	// Create an encrypted file
	encryptedFile, err := os.Create("encrypted.txt")
	if err != nil {
		log.Fatalf("Failed to create encrypted file: %v", err)
	}
	defer encryptedFile.Close()

	// Encrypt the message and write to the encrypted file
	armorWriter := armor.NewWriter(encryptedFile)
	recipient, err := age.ParseX25519Recipient(string(publicKey))
	if err != nil {
		log.Fatalf("Failed to parse recipient public key: %v", err)
	}
	w, err := age.Encrypt(armorWriter, recipient)
	if err != nil {
		log.Fatalf("Failed to create encryption writer: %v", err)
	}
	if _, err := io.WriteString(w, msg); err != nil {
		log.Fatalf("Failed to write to encrypted file: %v", err)
	}
	w.Close()
	armorWriter.Close()
}
