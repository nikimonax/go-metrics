package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"os"
)

const DefaultHashHeaderKey = "HashSHA256"

type SecurityConfig struct {
	Header        string
	HashingKey    string
	HashingFunc   func() hash.Hash
	CryptoKeyFile string
}

func readBlockPEM(filePath string) (*pem.Block, error) {
	pemBytes, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemBytes)

	if block == nil {
		return nil, errors.New("failed read PEM block")
	}

	return block, nil
}

func (cfg *SecurityConfig) ReadPrivateKey() (*rsa.PrivateKey, error) {
	block, err := readBlockPEM(cfg.CryptoKeyFile)

	if err != nil {
		return nil, err
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	privateKey, ok := parsedKey.(*rsa.PrivateKey)

	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return privateKey, nil
}

func (cfg *SecurityConfig) ReadPublicKey() (*rsa.PublicKey, error) {
	block, err := readBlockPEM(cfg.CryptoKeyFile)

	if err != nil {
		return nil, err
	}

	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	publicKey, ok := parsedKey.(*rsa.PublicKey)

	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return publicKey, nil
}
