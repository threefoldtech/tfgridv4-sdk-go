package client

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateFarm(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	keyPair, err := parseKeysFromMnemonicOrSeed(testMnemonic)
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

	request = newClientWithAccountNoNode
	c, err := NewRegistrarClient(baseURL, testMnemonic)
	require.NoError(err)

	stellarAddr := "GBB3H4F7N3R26I6XU2V2P5WYJZQ5N7E4FQH6E5D2X3OGJ2KLTGZXQW34"
	t.Run("test create farm with status conflict", func(t *testing.T) {
		request = createFarmStatusConflict
		_, err = c.CreateFarm(farm.FarmName, stellarAddr, farm.Dedicated)
		require.Error(err)
	})

	t.Run("test create farm with status ok", func(t *testing.T) {
		request = createFarmStatusCreated
		result, err := c.CreateFarm(farm.FarmName, stellarAddr, farm.Dedicated)
		require.NoError(err)
		require.Equal(farm.FarmID, result)
	})
}

func TestUpdateFarm(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	keyPair, err := parseKeysFromMnemonicOrSeed(testMnemonic)
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

	t.Run("test update farm with status unauthorzed", func(t *testing.T) {
		request = newClientWithNoAccount
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		request = updateFarmWithStatusUnauthorized
		name := "notFreeFarm"
		err = c.UpdateFarm(farmID, FarmUpdate{FarmName: &name})
		require.Error(err)
	})

	t.Run("test update farm with status ok", func(t *testing.T) {
		count = 0
		request = newClientWithAccountNoNode
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		request = updateFarmWithStatusOK
		name := "notFreeFarm"
		err = c.UpdateFarm(farmID, FarmUpdate{FarmName: &name})
		require.NoError(err)
	})
}

func TestGetFarm(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	keyPair, err := parseKeysFromMnemonicOrSeed(testMnemonic)
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

	t.Run("test get farm with status not found", func(t *testing.T) {
		request = getFarmWithStatusNotfound
		_, err = c.GetFarm(farmID)
		require.Error(err)
	})

	t.Run("test get farm with status ok", func(t *testing.T) {
		request = getFarmWithStatusOK
		result, err := c.GetFarm(farmID)
		require.NoError(err)
		require.Equal(result, farm)
	})
}

func TestApproveNodes(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	keyPair, err := parseKeysFromMnemonicOrSeed(testMnemonic)
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

	t.Run("test approve nodes with status unauthorized", func(t *testing.T) {
		count = 0
		request = newClientWithNoAccount
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		count = 0
		request = approveNodesWithStatusUnauthorized
		err = c.ApproveNodes(farmID, []uint64{1, 2, 3})
		require.Error(err)
	})

	t.Run("test approve nodes with status ok", func(t *testing.T) {
		count = 0
		request = newClientWithAccountNoNode
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		count = 0
		request = approveNodesWithStatusOK
		err = c.ApproveNodes(farmID, []uint64{1, 2, 3})
		require.NoError(err)
	})
}
