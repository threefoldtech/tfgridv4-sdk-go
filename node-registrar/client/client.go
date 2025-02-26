package client

import (
	"crypto/ed25519"
	"net/http"

	"github.com/pkg/errors"
)

type keyPair struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

type RegistrarClient struct {
	httpClient http.Client
	keyPair    keyPair
	nodeID     uint64
	twinID     uint64
	baseURL    string
}

func NewRegistrarClient(baseURL string, privateKey ed25519.PrivateKey) (cli RegistrarClient, err error) {
	client := http.DefaultClient

	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return cli, errors.Wrap(err, "failed to get public key of provided private key")
	}

	cli = RegistrarClient{
		httpClient: *client,
		keyPair:    keyPair{privateKey, publicKey},
		baseURL:    baseURL,
	}

	account, err := cli.GetAccountByPK(publicKey)
	if errors.Is(err, ErrorAccountNotFround) {
		return cli, nil
	} else if err != nil {
		return cli, errors.Wrap(err, "failed to get account with public key")
	}

	cli.twinID = account.TwinID
	node, err := cli.GetNodeByTwinID(account.TwinID)
	if errors.Is(err, ErrorNodeNotFround) {
		return cli, nil
	} else if err != nil {
		return cli, errors.Wrapf(err, "failed to get node with twin id %d", account.TwinID)
	}

	cli.nodeID = node.NodeID
	return
}
