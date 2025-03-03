package cmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

func UpdateAccount(seed string, network string, relays []string, rmbEncKey string) (err error) {
	u, ok := urls[network]
	if !ok {
		return fmt.Errorf("invalid network %s", network)
	}

	if len(seed) == 0 {
		return fmt.Errorf("can not initialize registrar client with no seed")
	}

	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return err
	}

	privateKey := ed25519.NewKeyFromSeed(seedBytes)
	cli, err := client.NewRegistrarClient(u, privateKey)
	if err != nil {
		return err
	}

	var opts []client.UpdateAccountOpts
	if len(relays) > 0 {
		opts = append(opts, client.UpdateAccountWithRelays(relays))
	}

	if len(rmbEncKey) > 0 {
		opts = append(opts, client.UpdateAccountWithRMBEncKey(rmbEncKey))
	}

	return cli.UpdateAccount(opts...)
}

func UpdateFarm(seed string, network string, farmID uint64, farmName string, dedicated bool) (err error) {
	u, ok := urls[network]
	if !ok {
		return fmt.Errorf("invalid network %s", network)
	}

	if len(seed) == 0 {
		return fmt.Errorf("can not initialize registrar client with no seed")
	}

	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return err
	}

	privateKey := ed25519.NewKeyFromSeed(seedBytes)
	cli, err := client.NewRegistrarClient(u, privateKey)
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

	return cli.UpdateFarm(farmID, opts...)
}

func UpdateVersion(seed string, network string, version string, safeToUpgrade bool) (err error) {
	u, ok := urls[network]
	if !ok {
		return fmt.Errorf("invalid network %s", network)
	}

	if len(seed) == 0 {
		return fmt.Errorf("can not initialize registrar client with no seed")
	}

	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return err
	}

	privateKey := ed25519.NewKeyFromSeed(seedBytes)
	cli, err := client.NewRegistrarClient(u, privateKey)
	if err != nil {
		return err
	}

	return cli.SetZosVersion(version, safeToUpgrade)
}
