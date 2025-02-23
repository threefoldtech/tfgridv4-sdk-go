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

	pk, seed, err := aliceKeys()
	require.NoError(err)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(pk)
	account.PublicKey = publicKeyBase64

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, body := accountHandler(r, request, count, require)
		w.WriteHeader(statusCode)
		_, err := w.Write(body)
		require.NoError(err)
		count++
	}))
	defer testServer.Close()

	baseURL, err := url.JoinPath(testServer.URL, "v1")
	require.NoError(err)

	count = 0
	request = newClientWithNoAccount
	c, err := NewRegistrarClient(baseURL, seed)
	require.NoError(err)

	t.Run("test create account created successfully", func(t *testing.T) {
		count = 0
		request = createAccountStatusCreated
		result, err := c.CreateAccount(account.Relays, account.RMBEncKey)
		require.NoError(err)
		require.Equal(account, result)
	})
}

func TestUpdateAccount(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	pk, seed, err := aliceKeys()
	require.NoError(err)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(pk)
	account.PublicKey = publicKeyBase64

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, body := accountHandler(r, request, count, require)
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
		c, err := NewRegistrarClient(baseURL, seed)

		require.NoError(err)
		require.Equal(c.twinID, account.TwinID)
		require.Equal([]byte(c.keyPair.publicKey), pk)

		count = 0
		request = updateAccountWithStatusOK
		relays := []string{"relay1"}
		err = c.UpdateAccount(UpdateAccountWithRelays(relays))
		require.NoError(err)
	})

	t.Run("test update account account not found", func(t *testing.T) {
		count = 0
		request = newClientWithNoAccount
		c, err := NewRegistrarClient(baseURL, seed)

		require.NoError(err)
		require.Equal([]byte(c.keyPair.publicKey), pk)

		count = 0
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

	pk, seed, err := aliceKeys()
	require.NoError(err)
	publicKeyBase64 := base64.StdEncoding.EncodeToString(pk)
	account.PublicKey = publicKeyBase64

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode, body := accountHandler(r, request, count, require)
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
	c, err := NewRegistrarClient(baseURL, seed)
	require.NoError(err)
	require.Equal(c.twinID, account.TwinID)
	require.Equal([]byte(c.keyPair.publicKey), pk)

	t.Run("test get account with id account not found", func(t *testing.T) {
		count = 0
		request = getAccountWithIDStatusNotFount
		_, err := c.GetAccount(account.TwinID)
		require.Error(err)
	})

	t.Run("test update account account not found", func(t *testing.T) {
		count = 0
		request = getAccountWithIDStatusOK
		acc, err := c.GetAccount(account.TwinID)
		require.NoError(err)
		require.Equal(account, acc)
	})
}
