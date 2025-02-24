package client

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRegistrarClient(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	pk, seed, publicKeyBase64, err := aliceKeys()
	require.NoError(err)
	account.PublicKey = publicKeyBase64

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

	t.Run("test new registrar client with no account", func(t *testing.T) {
		count = 0
		request = newClientWithNoAccount
		c, err := NewRegistrarClient(baseURL, seed)
		require.NoError(err)
		require.Equal(uint64(0), c.twinID)
		require.Equal(uint64(0), c.nodeID)
		require.Equal([]byte(c.keyPair.publicKey), pk)
	})

	t.Run("test new registrar client with account and no node", func(t *testing.T) {
		count = 0
		request = newClientWithAccountNoNode
		c, err := NewRegistrarClient(baseURL, seed)
		require.NoError(err)
		require.Equal(account.TwinID, c.twinID)
		require.Equal(uint64(0), c.nodeID)
		require.Equal(pk, []byte(c.keyPair.publicKey))
	})
	t.Run("test new registrar client with account and node", func(t *testing.T) {
		count = 0
		request = newClientWithAccountAndNode
		c, err := NewRegistrarClient(baseURL, seed)
		require.NoError(err)
		require.Equal(account.TwinID, c.twinID)
		require.Equal(nodeID, c.nodeID)
		require.Equal(pk, []byte(c.keyPair.publicKey))
	})
}
