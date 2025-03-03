package client

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegistarNode(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	// publicKey, privateKey, err := aliceKeys()
	// require.NoError(err)
	// account.PublicKey = base64.StdEncoding.EncodeToString(publicKey)
	//
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

	t.Run("test registar node no account", func(t *testing.T) {
		request = newClientWithNoAccount
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		request = registerNodeWithNoAccount
		_, err = c.RegisterNode(farmID, twinID, []Interface{}, Location{}, Resources{}, "", false, false)
		require.Error(err)
	})

	t.Run("test registar node, node already exist", func(t *testing.T) {
		count = 0
		request = newClientWithAccountAndNode
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		count = 0
		request = registerNodeStatusConflict
		_, err = c.RegisterNode(farmID, twinID, []Interface{}, Location{}, Resources{}, "", false, false)
		require.Error(err)
	})

	t.Run("test registar node, created successfully", func(t *testing.T) {
		count = 0
		request = newClientWithAccountNoNode
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		count = 0
		request = registerNodeStatusCreated
		result, err := c.RegisterNode(farmID, twinID, []Interface{}, Location{}, Resources{}, "", false, false)
		require.NoError(err)
		require.Equal(nodeID, result)
	})
}

func TestUpdateNode(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	// publicKey, privateKey, err := aliceKeys()
	// require.NoError(err)
	// account.PublicKey = base64.StdEncoding.EncodeToString(publicKey)
	//
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

	request = newClientWithAccountAndNode
	c, err := NewRegistrarClient(baseURL, testMnemonic)
	require.NoError(err)

	t.Run("test update node with status ok", func(t *testing.T) {
		count = 0
		request = updateNodeStatusOK
		err = c.UpdateNode(UpdateNodesWithFarmID(2))
		require.NoError(err)
	})

	t.Run("test update node uptime", func(t *testing.T) {
		request = updateNodeSendUptimeReport

		report := UptimeReport{
			Uptime:    40 * time.Minute,
			Timestamp: time.Now(),
		}
		err = c.ReportUptime(report)
		require.NoError(err)
	})
}

func TestGetNode(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

	// publicKey, privateKey, err := aliceKeys()
	// require.NoError(err)
	// account.PublicKey = base64.StdEncoding.EncodeToString(publicKey)

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

	request = newClientWithAccountAndNode
	c, err := NewRegistrarClient(baseURL, testMnemonic)
	require.NoError(err)

	t.Run("test get node status not found", func(t *testing.T) {
		request = getNodeWithIDStatusNotFound
		_, err = c.GetNode(nodeID)
		require.Error(err)
	})

	t.Run("test get node, status ok", func(t *testing.T) {
		request = getNodeWithIDStatusOK
		result, err := c.GetNode(nodeID)
		require.NoError(err)
		require.Equal(node, result)
	})

	t.Run("test get node with twin id", func(t *testing.T) {
		request = getNodeWithTwinID
		result, err := c.GetNodeByTwinID(twinID)
		require.NoError(err)
		require.Equal(node, result)
	})

	t.Run("test list nodes of specific farm", func(t *testing.T) {
		request = listNodesInFarm
		result, err := c.ListNodes(ListNodesWithFarmID(farmID))
		require.NoError(err)
		require.Equal([]Node{node}, result)
	})
}
