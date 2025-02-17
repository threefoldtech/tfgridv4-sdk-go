package client

import (
	"fmt"
)

var ErrorAccountNotFround = fmt.Errorf("failed to get requested account from node regiatrar")

func (c RegistrarClient) CreateAccount(relays []string, rmbEncKey string) (account Account, err error) {
	return c.createTwin(relays, rmbEncKey)
}

func (c RegistrarClient) UpdateAccount(relays []string, rmbEncKey string) (err error) {
	return
}

func (c RegistrarClient) EnsureAccount(pk []byte) (account Account, err error) {
	return
}

func (c RegistrarClient) GetAccount(id uint64) (account Account, err error) {
	return
}

func (c RegistrarClient) GetAccountByPK(pk []byte) (account Account, err error) {
	return
}

func (c *RegistrarClient) createTwin(relays []string, rmbEncKey string) (result Account, err error) {
	return
}
