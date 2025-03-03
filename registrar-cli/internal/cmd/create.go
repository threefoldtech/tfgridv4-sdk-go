package cmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
	"github.com/threefoldtech/tfgrid4-sdk-go/registrar-cli/internal/seedgen"
)

var urls = map[string]string{
	"dev":  "https://registrar.dev4.grid.tf/v1",
	"qa":   "https://registrar.qa4.grid.tf/v1",
	"test": "https://registrar.test4.grid.tf/v1",
	"main": "https://registrar.prod4.grid.tf/v1",
}

func CreaeteAccount(seed string, network string, relays []string, rmbEncKey string) (account client.Account, err error) {
	u, ok := urls[network]
	if !ok {
		return account, fmt.Errorf("invalid network %s", network)
	}

	if len(seed) == 0 {
		seed, err = seedgen.GenerateRandomSeed()
		if err != nil {
			return
		}
		log.Info().Msgf("New Seed (Hex): %s", seed)
	}

	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return
	}

	privateKey := ed25519.NewKeyFromSeed(seedBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	log.Info().Msgf("public key (Hex): %s", hex.EncodeToString(publicKey))

	cli, err := client.NewRegistrarClient(u, privateKey)
	if err != nil {
		return
	}

	return cli.CreateAccount(relays, rmbEncKey)
}

func CreaeteFarm(seed string, network string, farmName string, dedicated bool) (farmID uint64, err error) {
	u, ok := urls[network]
	if !ok {
		return farmID, fmt.Errorf("invalid network %s", network)
	}

	if len(seed) == 0 {
		return farmID, fmt.Errorf("can not initialize registrar client with no seed")
	}

	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return
	}

	privateKey := ed25519.NewKeyFromSeed(seedBytes)
	cli, err := client.NewRegistrarClient(u, privateKey)
	if err != nil {
		return
	}

	return cli.CreateFarm(farmName, dedicated)
}
