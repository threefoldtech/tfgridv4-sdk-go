package client

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	_, keyPair, err := parseKeysFromMnemonicOrSeed(testMnemonic)
	require.NoError(err)
	account.PublicKey = base64.StdEncoding.EncodeToString(keyPair.Public())

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, body := serverHandler(r, request, count, require)
		w.WriteHeader(statusCode)
		_, err := w.Write(body)
		require.NoError(err)
		count++
	}))
	defer testServer.Close()

	baseURL, err := url.JoinPath(testServer.URL, "v1")
	require.NoError(err)

	request = newClientWithNoAccount
	c, err := NewRegistrarClient(baseURL, testMnemonic)
	require.NoError(err)

	t.Run("test create account created successfully", func(t *testing.T) {
		request = createAccountStatusCreated
		result, _, err := c.CreateAccount(account.Relays, account.RMBEncKey)
		require.NoError(err)
		require.Equal(account, result)
	})
}

func TestUpdateAccount(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	_, keyPair, err := parseKeysFromMnemonicOrSeed(testMnemonic)
	require.NoError(err)
	account.PublicKey = base64.StdEncoding.EncodeToString(keyPair.Public())

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, body := serverHandler(r, request, count, require)
		w.WriteHeader(statusCode)
		_, err := w.Write(body)
		require.NoError(err)

		count++
	}))
	defer testServer.Close()

	baseURL, err := url.JoinPath(testServer.URL, "v1")
	require.NoError(err)

	t.Run("test update account updated successfully", func(t *testing.T) {
		count = 0
		request = newClientWithAccountNoNode
		c, err := NewRegistrarClient(baseURL, testMnemonic)

		require.NoError(err)
		require.Equal(c.twinID, account.TwinID)
		require.Equal(c.keyPair, keyPair)

		request = updateAccountWithStatusOK
		relays := []string{"relay1"}
		err = c.UpdateAccount(UpdateAccountWithRelays(relays))
		require.NoError(err)
	})

	t.Run("test update account account not found", func(t *testing.T) {
		request = newClientWithNoAccount
		c, err := NewRegistrarClient(baseURL, testMnemonic)

		require.NoError(err)
		require.Equal(c.keyPair, keyPair)

		request = updateAccountWithNoAccount
		relays := []string{"relay1"}
		err = c.UpdateAccount(UpdateAccountWithRelays(relays))
		require.Error(err)
	})
}

func TestGetAccount(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	_, keyPair, err := parseKeysFromMnemonicOrSeed(testMnemonic)
	require.NoError(err)
	account.PublicKey = base64.StdEncoding.EncodeToString(keyPair.Public())

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, body := serverHandler(r, request, count, require)
		w.WriteHeader(statusCode)
		_, err := w.Write(body)
		require.NoError(err)
		count++
	}))
	defer testServer.Close()

	baseURL, err := url.JoinPath(testServer.URL, "v1")
	require.NoError(err)

	count = 0
	request = newClientWithAccountNoNode
	c, err := NewRegistrarClient(baseURL, testMnemonic)
	require.NoError(err)
	require.Equal(account.TwinID, c.twinID)
	require.Equal(keyPair, c.keyPair)

	t.Run("test get account with id account not found", func(t *testing.T) {
		request = getAccountWithIDStatusNotFount
		_, err := c.GetAccount(account.TwinID)
		require.Error(err)
	})

	t.Run("test get account account not found", func(t *testing.T) {
		request = getAccountWithIDStatusOK
		acc, err := c.GetAccount(account.TwinID)
		require.NoError(err)
		require.Equal(account, acc)
	})
}
