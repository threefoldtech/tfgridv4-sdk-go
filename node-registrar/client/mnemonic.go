package client

import (
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	subkeyEd25519 "github.com/vedhavyas/go-subkey/v2/ed25519"

	"github.com/vedhavyas/go-subkey/v2"
)

func (c *RegistrarClient) Mnemonic() string {
	return c.mnemonic
}

func parseKeysFromMnemonicOrSeed(mnemonicOrSeed string) (keypair subkey.KeyPair, err error) {
	// otherwise drive key pair from seed
	keypair, err = subkey.DeriveKeyPair(subkeyEd25519.Scheme{}, mnemonicOrSeed)
	if err != nil {
		return keypair, errors.Wrapf(err, "Failed to derive key pair from seed %s", mnemonicOrSeed)
	}

	return keypair, nil
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
