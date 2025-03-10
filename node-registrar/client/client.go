package client

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/vedhavyas/go-subkey"
)

type RegistrarClient struct {
	httpClient http.Client
	baseURL    string
	keyPair    subkey.KeyPair
	mnemonic   string
	nodeID     uint64
	twinID     uint64
}

func NewRegistrarClient(baseURL string, mnemonicOrSeed ...string) (cli RegistrarClient, err error) {
	client := http.DefaultClient

	cli = RegistrarClient{
		httpClient: *client,
		baseURL:    baseURL,
	}

	if len(mnemonicOrSeed) == 0 {
		return cli, nil
	}

	keyPair, err := parseKeysFromMnemonicOrSeed(mnemonicOrSeed[0])
	if err != nil {
		return cli, errors.Wrapf(err, "Failed to derive key pair from mnemonic/seed phrase %s", mnemonicOrSeed[0])
	}

	cli.keyPair = keyPair
	cli.mnemonic = mnemonicOrSeed[0]

	account, err := cli.GetAccountByPK(keyPair.Public())
	if errors.Is(err, ErrorAccountNotFound) {
		return cli, nil
	} else if err != nil {
		return cli, errors.Wrap(err, "failed to get account with public key")
	}

	cli.twinID = account.TwinID
	node, err := cli.GetNodeByTwinID(account.TwinID)
	if errors.Is(err, ErrorNodeNotFound) {
		return cli, nil
	} else if err != nil {
		return cli, errors.Wrapf(err, "failed to get node with twin id %d", account.TwinID)
	}

	cli.nodeID = node.NodeID
	return
}
