package cmd

import (
	"fmt"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

func CreateFarm(mnemonic, network, farmName, stellarAddress string, dedicated bool) (farmID uint64, err error) {
	u, ok := urls[network]
	if !ok {
		return farmID, fmt.Errorf("invalid network %s", network)
	}

	if len(mnemonic) == 0 {
		return farmID, fmt.Errorf("can not initialize registrar client with no mnemonic")
	}

	cli, err := client.NewRegistrarClient(u, mnemonic)
	if err != nil {
		return
	}

	return cli.CreateFarm(farmName, stellarAddress, dedicated)
}

func GetFarm(network string, farmID uint64) (farm client.Farm, err error) {
	u, ok := urls[network]
	if !ok {
		return farm, fmt.Errorf("invalid network %s", network)
	}

	cli, err := client.NewRegistrarClient(u)
	if err != nil {
		return
	}

	return cli.GetFarm(farmID)
}

func UpdateFarm(farmID uint64, mnemonic, network, farmName, stellarAddress string, dedicated bool) (err error) {
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

	var opts []client.UpdateFarmOpts
	if len(farmName) > 0 {
		opts = append(opts, client.UpdateFarmWithName(farmName))
	}

	if dedicated {
		opts = append(opts, client.UpdateFarmWithDedicated())
	}
	if len(stellarAddress) != 0 {
		// opts = append(opts, client.Updfarm)
	}

	return cli.UpdateFarm(farmID, opts...)
}
