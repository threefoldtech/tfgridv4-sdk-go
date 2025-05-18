package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	subkey "github.com/vedhavyas/go-subkey/v2"
)

var ErrorAccountNotFound = fmt.Errorf("failed to get requested account from node registrar")

// CreateAccount create new account on the registrar with uniqe mnemonic.
func (c *RegistrarClient) CreateAccount(relays []string, rmbEncKey string) (account Account, mnemonic string, err error) {
	return c.createAccount(relays, rmbEncKey)
}

// GetAccount get an account using either its twinID
func (c *RegistrarClient) GetAccount(twinID uint64) (account Account, err error) {
	return c.getAccount(twinID)
}

// GetAccountByPK get an account using either its its publicKey.
func (c *RegistrarClient) GetAccountByPK(publicKey []byte) (account Account, err error) {
	return c.getAccountByPK(publicKey)
}

// UpdateAccount update the account configuration (relays or rmbEncKey).
func (c *RegistrarClient) UpdateAccount(opts ...UpdateAccountOpts) (err error) {
	return c.updateAccount(opts)
}

type accountCfg struct {
	relays    []string
	rmbEncKey string
}

type (
	UpdateAccountOpts func(*accountCfg)
)

// UpdateAccountWithRelays update the account relays
func UpdateAccountWithRelays(relays []string) UpdateAccountOpts {
	return func(n *accountCfg) {
		n.relays = relays
	}
}

// UpdateAccountWithRMBEncKey update the account rmb encryption key
func UpdateAccountWithRMBEncKey(rmbEncKey string) UpdateAccountOpts {
	return func(n *accountCfg) {
		n.rmbEncKey = rmbEncKey
	}
}

// EnsureAccount ensures that an account is created with specific seed/mnemonic.
func (c *RegistrarClient) EnsureAccount(relays []string, rmbEncKey string) (account Account, err error) {
	return c.ensureAccount(relays, rmbEncKey)
}

func (c *RegistrarClient) createAccount(relays []string, rmbEncKey string) (account Account, mnemonic string, err error) {
	url, err := url.JoinPath(c.baseURL, "accounts")
	if err != nil {
		return account, mnemonic, errors.Wrap(err, "failed to construct registrar url")
	}

	var keyPair subkey.KeyPair
	if len(c.mnemonic) != 0 {
		mnemonic = c.mnemonic
		keyPair, err = parseKeysFromMnemonicOrSeed(c.mnemonic)
	} else {
		mnemonic, keyPair, err = generateNewMnemonic()
	}
	if err != nil {
		return account, mnemonic, err
	}

	c.keyPair = keyPair
	c.mnemonic = mnemonic

	publicKeyBase64 := base64.StdEncoding.EncodeToString(c.keyPair.Public())

	timestamp := time.Now().Unix()
	challenge := []byte(fmt.Sprintf("%d:%v", timestamp, publicKeyBase64))
	signature, err := keyPair.Sign(challenge)
	if err != nil {
		return account, mnemonic, errors.Wrap(err, "failed to sign account creation request")
	}

	data := map[string]any{
		"public_key":  base64.StdEncoding.EncodeToString(c.keyPair.Public()),
		"signature":   base64.StdEncoding.EncodeToString(signature),
		"timestamp":   timestamp,
		"rmb_enc_key": rmbEncKey,
		"relays":      relays,
	}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(data)
	if err != nil {
		return account, mnemonic, errors.Wrap(err, "failed to parse request body")
	}

	resp, err := c.httpClient.Post(url, "application/json", &body)
	if err != nil {
		return account, mnemonic, errors.Wrap(err, "failed to send request to the registrar")
	}

	if resp == nil {
		return account, mnemonic, errors.New("failed to create account, no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		err = parseResponseError(resp.Body)
		return account, mnemonic, errors.Wrapf(err, "failed to create account with status %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&account)

	c.twinID = account.TwinID
	return
}

func (c *RegistrarClient) getAccount(id uint64) (account Account, err error) {
	url, err := url.JoinPath(c.baseURL, "accounts")
	if err != nil {
		return account, errors.Wrap(err, "failed to construct registrar url")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("twin_id", fmt.Sprint(id))
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}

	if resp == nil {
		return account, errors.New("failed to get account, no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return account, ErrorAccountNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return account, errors.Wrapf(err, "failed to get account by twin id with status code %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&account)
	return
}

func (c *RegistrarClient) getAccountByPK(pk []byte) (account Account, err error) {
	url, err := url.JoinPath(c.baseURL, "accounts")
	if err != nil {
		return account, errors.Wrap(err, "failed to construct registrar url")
	}

	publicKeyBase64 := base64.StdEncoding.EncodeToString(pk)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return account, err
	}

	q := req.URL.Query()
	q.Add("public_key", publicKeyBase64)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return account, err
	}

	if resp == nil {
		return account, errors.New("no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return account, ErrorAccountNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return account, errors.Wrapf(err, "failed to get account by public_key with status code %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&account)

	return account, err
}

func (c *RegistrarClient) updateAccount(opts []UpdateAccountOpts) (err error) {
	err = c.ensureTwinID()
	if err != nil {
		return errors.Wrap(err, "failed to ensure twin id")
	}
	url, err := url.JoinPath(c.baseURL, "accounts", fmt.Sprint(c.twinID))
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	var body bytes.Buffer
	data := parseUpdateAccountOpts(opts)

	err = json.NewEncoder(&body).Encode(data)
	if err != nil {
		return errors.Wrap(err, "failed to parse request body")
	}

	req, err := http.NewRequest("PATCH", url, &body)
	if err != nil {
		return
	}

	authHeader, err := c.signRequest(time.Now().Unix())
	if err != nil {
		return errors.Wrap(err, "failed to sign request")
	}
	req.Header.Set("X-Auth", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}

	if resp == nil {
		return errors.New("failed to update account, no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseResponseError(resp.Body)
	}

	return
}

func (c *RegistrarClient) ensureAccount(relays []string, rmbEncKey string) (account Account, err error) {
	account, err = c.GetAccountByPK(c.keyPair.Public())
	if errors.Is(err, ErrorAccountNotFound) {
		account, _, err = c.CreateAccount(relays, rmbEncKey)
	}
	return account, err
}

// ensureTwinID ensures that the RegistrarClient is set up properly with a valid public key representing an account on the registrar
func (c *RegistrarClient) ensureTwinID() error {
	if c.twinID != 0 {
		return nil
	}

	twin, err := c.getAccountByPK(c.keyPair.Public())
	if err != nil {
		return errors.Wrap(err, "failed to get the account of the node, registrar client was not set up properly")
	}

	c.twinID = twin.TwinID
	return nil
}

func parseUpdateAccountOpts(opts []UpdateAccountOpts) map[string]any {
	cfg := accountCfg{
		rmbEncKey: "",
		relays:    []string{},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	data := map[string]any{}

	if len(cfg.relays) != 0 {
		data["relays"] = cfg.relays
	}

	if len(cfg.rmbEncKey) != 0 {
		data["rmb_enc_key"] = cfg.rmbEncKey
	}

	return data
}
