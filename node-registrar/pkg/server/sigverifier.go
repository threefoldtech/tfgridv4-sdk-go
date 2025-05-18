package server

import (
	"github.com/pkg/errors"
	subkey "github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/ed25519"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

const PubKeySize = 32 // ED25519/SR25519 public key size

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrVerifyFailed = errors.New("signature verification failed")
)

func verifySignature(publicKey, challenge, signature []byte) error {
	// Verify public key length
	if len(publicKey) != PubKeySize {
		return errors.Wrapf(ErrInvalidInput, "invalid public key size: expected %d, got %d", PubKeySize, len(publicKey))
	}

	if len(challenge) == 0 {
		return errors.Wrap(ErrInvalidInput, "invalid challenge size, not expected to be zero")
	}

	if len(signature) == 0 {
		return errors.Wrap(ErrInvalidInput, "invalid signature size, not expected to be zero")
	}

	// Try ED25519 verification first
	if verifyWithScheme(ed25519.Scheme{}, publicKey, challenge, signature) ||
		// Fallback to SR25519 verification
		verifyWithScheme(sr25519.Scheme{}, publicKey, challenge, signature) {
		return nil
	}

	return ErrVerifyFailed
}

func verifyWithScheme(scheme subkey.Scheme, publicKey, challenge, signature []byte) bool {
	key, err := scheme.FromPublicKey(publicKey)
	if err != nil {
		return false
	}
	return key.Verify(challenge, signature)
}
