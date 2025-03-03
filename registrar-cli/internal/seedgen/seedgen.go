package seedgen

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomSeed() (string, error) {
	s := make([]byte, 32)
	_, err := rand.Read(s)
	if err != nil {
		return "", err
	}

	seed := hex.EncodeToString(s)
	return seed, nil
}

func GenerateRandomKey() ([]byte, error) {
	seed, err := GenerateRandomSeed()
	if err != nil {
		return nil, err
	}

	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return nil, err
	}

	privateKey := ed25519.NewKeyFromSeed(seedBytes)
	return privateKey, nil
}
