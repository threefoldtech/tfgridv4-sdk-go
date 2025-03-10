package cmd

import (
	"fmt"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

func GetVersion(network string) (version client.ZosVersion, err error) {
	u, ok := urls[network]
	if !ok {
		return version, fmt.Errorf("invalid network %s", network)
	}

	cli, err := client.NewRegistrarClient(u)
	if err != nil {
		return
	}

	return cli.GetZosVersion()
}

func UpdateVersion(mnemonic string, network string, version string, safeToUpgrade bool) (err error) {
	u, ok := urls[network]
	if !ok {
		return fmt.Errorf("invalid network %s", network)
	}

	if len(mnemonic) == 0 {
		return fmt.Errorf("can not initialize registrar client with no mnemonic")
	}

	cli, err := client.NewRegistrarClient(u, mnemonic)
	if err != nil {
		return err
	}

	return cli.SetZosVersion(version, safeToUpgrade)
}
