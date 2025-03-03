package client

import (
	"encoding/hex"

	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	subkeyEd25519 "github.com/vedhavyas/go-subkey/v2/ed25519"

	"github.com/vedhavyas/go-subkey/v2"
)

func (c RegistrarClient) Mnemonic() string {
	return c.mnemonic
}

func parseKeysFromMnemonicOrSeed(mnemonicOrSeed string) (mnemonic string, keypair subkey.KeyPair, err error) {
	if ok := bip39.IsMnemonicValid(mnemonicOrSeed); ok {
		// If mnemonic is valid drive key pair from mnemonic
		keypair, err = subkey.DeriveKeyPair(subkeyEd25519.Scheme{}, mnemonicOrSeed)
		if err != nil {
			return "", keypair, errors.Wrapf(err, "Failed to derive key pair from mnemonic phrase %s", mnemonicOrSeed)
		}
		return mnemonicOrSeed, keypair, nil
	}

	// otherwise parse it as seed
	seed, err := hex.DecodeString(mnemonicOrSeed)
	if err != nil {
		return "", keypair, errors.Wrap(err, "failed to decode seed")
	}

	mnemonic, err = bip39.NewMnemonic(seed)
	if err != nil {
		return "", keypair, errors.Wrapf(err, "Failed to generate mnemonic from %s", mnemonicOrSeed)
	}

	keypair, err = subkey.DeriveKeyPair(subkeyEd25519.Scheme{}, mnemonic)
	if err != nil {
		return "", keypair, errors.Wrapf(err, "Failed to derive key pair from seed %s", mnemonicOrSeed)
	}

	return
}

func generateNewMnemonic() (mnemonic string, keypair subkey.KeyPair, err error) {
	// Generate 128-bit entropy (12-word mnemonic)
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return mnemonic, keypair, errors.Wrap(err, "Failed to generate entropy")
	}

	// Generate mnemonic from entropy
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return mnemonic, keypair, errors.Wrap(err, "Failed to generate mnemonic")
	}

	// Drive key pair from mnemonic
	keypair, err = subkey.DeriveKeyPair(subkeyEd25519.Scheme{}, mnemonic)
	if err != nil {
		return mnemonic, keypair, errors.Wrapf(err, "Failed to derive key pair from mnemonic phrase %s", mnemonic)
	}
	return
}
