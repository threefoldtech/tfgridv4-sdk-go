package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

var urls = map[string]string{
	"dev":  "https://registrar.dev4.grid.tf/v1",
	"qa":   "https://registrar.qa4.grid.tf/v1",
	"test": "https://registrar.test4.grid.tf/v1",
	"main": "https://registrar.prod4.grid.tf/v1",
}

func CreateAccount(network string, relays []string, rmbEncKey string, mnemonicOrSeed ...string) (account client.Account, mnemonic string, err error) {
	u, ok := urls[network]
	if !ok {
		return account, "", fmt.Errorf("invalid network %s", network)
	}

	var cli client.RegistrarClient
	cli, err = client.NewRegistrarClient(u, mnemonicOrSeed...)
	if err != nil {
		return
	}

	return cli.CreateAccount(relays, rmbEncKey)
}

func GetAccount(network string, twinID uint64, pk string) (account client.Account, err error) {
	u, ok := urls[network]
	if !ok {
		return account, fmt.Errorf("invalid network %s", network)
	}

	cli, err := client.NewRegistrarClient(u)
	if err != nil {
		return
	}

	if twinID != 0 {
		return cli.GetAccount(twinID)
	} else if len(pk) != 0 {

		publicKey, err := hex.DecodeString(pk)
		if err != nil {
			return account, err
		}

		return cli.GetAccountByPK(publicKey)
	}

	return account, fmt.Errorf("you need to provide either twin id or public key to load an account")
}

func UpdateAccount(mnemonic string, network string, relays []string, rmbEncKey string) (err error) {
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

	var opts []client.UpdateAccountOpts
	if len(relays) > 0 {
		opts = append(opts, client.UpdateAccountWithRelays(relays))
	}

	if len(rmbEncKey) > 0 {
		opts = append(opts, client.UpdateAccountWithRMBEncKey(rmbEncKey))
	}

	return cli.UpdateAccount(opts...)
}
