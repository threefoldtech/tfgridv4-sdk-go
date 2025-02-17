package client

import (
	"crypto/ed25519"
	"net/http"

	"github.com/pkg/errors"
)

type RegistrarClient struct {
	httpClient http.Client
	privateKey ed25519.PrivateKey
	nodeID     uint64
	twinID     uint64
	baseURL    string
}

func NewRegistrarClient(baseURL string, privateKey []byte) (cli RegistrarClient, err error) {
	client := http.DefaultClient

	sk := ed25519.NewKeyFromSeed(privateKey)
	pk, ok := sk.Public().(ed25519.PublicKey)
	if !ok {
		return cli, errors.Wrap(err, "failed to get public key of provided private key")
	}

	cli = RegistrarClient{
		httpClient: *client,
		privateKey: privateKey,
		baseURL:    baseURL,
	}

	account, err := cli.GetAccountByPK(pk)
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
