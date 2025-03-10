package cmd

import (
	"fmt"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

func GetNode(network string, nodeID, twinID uint64) (node client.Node, err error) {
	u, ok := urls[network]
	if !ok {
		return node, fmt.Errorf("invalid network %s", network)
	}

	cli, err := client.NewRegistrarClient(u)
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
