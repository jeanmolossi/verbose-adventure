package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jeanmolossi/verbose-adventure/internal/config"
)

var secretValue string

func main() {
	flag.StringVar(&secretValue, "secret", "", "-secret=<client-secret>")

	flag.Parse()

	if secretValue == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	rawSecret := []byte(secretValue)

	secret, err := encryptSecret(rawSecret, cfg.EncryptionKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("hex secret: " + base64.StdEncoding.EncodeToString(secret))
}

// encryptSecret aplica AES-GCM com chave base64 (32 bytes) e retorna nonce|ciphertext
func encryptSecret(plaintext []byte, keyB64 string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, fmt.Errorf("can not decode encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return append(nonce, gcm.Seal(nil, nonce, plaintext, nil)...), nil
}
