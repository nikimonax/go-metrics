package config

import (
	"errors"
	"hash"
	"os"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

const DefaultHashHeaderKey = "HashSHA256"

type SecurityConfig struct {
	Header        string
	HashingKey    string
	HashingFunc   func() hash.Hash
	CryptoKeyFile string
}

func (cfg *SecurityConfig) ReadPrivateKey() (*crypto.Key, error) {
	key, err := cfg.readKey()

	if err != nil {
		return nil, err
	}

	if !key.IsPrivate() {
		return nil, errors.New("not a private key")
	}

	unlocked, err := key.IsUnlocked()

	if err != nil {
		return nil, err
	}

	if !unlocked {
		return nil, errors.New("private key is locked")
	}

	return key, nil
}

func (cfg *SecurityConfig) ReadPublicKey() (*crypto.Key, error) {
	key, err := cfg.readKey()

	if err != nil {
		return nil, err
	}

	if key.IsPrivate() {
		return nil, errors.New("not a public key")
	}

	return key, nil
}

func (cfg *SecurityConfig) readKey() (*crypto.Key, error) {
	content, err := os.ReadFile(cfg.CryptoKeyFile)

	if err != nil {
		return nil, err
	}

	return crypto.NewKeyFromArmored(string(content))
}
