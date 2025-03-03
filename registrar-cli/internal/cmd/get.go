package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/seedgen"
)

func GetAccount(network string, twinID uint64, pk string) (account client.Account, err error) {
	u, ok := urls[network]
	if !ok {
		return account, fmt.Errorf("invalid network %s", network)
	}

	privateKey, err := seedgen.GenerateRandomKey()
	if err != nil {
		return
	}

	publicKey, err := hex.DecodeString(pk)
	if err != nil {
		return
	}

	cli, err := client.NewRegistrarClient(u, privateKey)
	if err != nil {
		return
	}

	if twinID != 0 {
		return cli.GetAccount(twinID)
	} else if len(publicKey) != 0 {
		return cli.GetAccountByPK(publicKey)
	}

	return account, fmt.Errorf("you need to provide either twin id or public key to load an account")
}

func GetFarm(network string, farmID uint64) (farm client.Farm, err error) {
	u, ok := urls[network]
	if !ok {
		return farm, fmt.Errorf("invalid network %s", network)
	}

	privateKey, err := seedgen.GenerateRandomKey()
	if err != nil {
		return
	}

	cli, err := client.NewRegistrarClient(u, privateKey)
	if err != nil {
		return
	}

	return cli.GetFarm(farmID)
}

func GetNode(network string, nodeID, twinID uint64) (node client.Node, err error) {
	u, ok := urls[network]
	if !ok {
		return node, fmt.Errorf("invalid network %s", network)
	}

	privateKey, err := seedgen.GenerateRandomKey()
	if err != nil {
		return
	}

	cli, err := client.NewRegistrarClient(u, privateKey)
	if err != nil {
		return
	}

	if nodeID != 0 {
		node, err = cli.GetNode(nodeID)
		if err != nil {
			return
		}

	} else if twinID != 0 {
		node, err = cli.GetNodeByTwinID(twinID)
		if err != nil {
			return
		}
	} else {
		return node, fmt.Errorf("you need to provide either twin id or node id to load a node")
	}

	return
}

func GetVersion(network string) (version client.ZosVersion, err error) {
	u, ok := urls[network]
	if !ok {
		return version, fmt.Errorf("invalid network %s", network)
	}

	privateKey, err := seedgen.GenerateRandomKey()
	if err != nil {
		return
	}

	cli, err := client.NewRegistrarClient(u, privateKey)
	if err != nil {
		return
	}

	return cli.GetZosVersion()
}
