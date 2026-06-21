package client

import (
	"encoding/base64"
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

	node := Node{
		TwinID: twinID,
		FarmID: farmID,
	}

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

	t.Run("test registar node no account", func(t *testing.T) {
		request = newClientWithNoAccount
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		request = registerNodeWithNoAccount
		_, err = c.RegisterNode(node)
		require.Error(err)
	})

	t.Run("test registar node, node already exist", func(t *testing.T) {
		count = 0
		request = newClientWithAccountAndNode
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		count = 0
		request = registerNodeStatusConflict
		_, err = c.RegisterNode(node)
		require.Error(err)
	})

	t.Run("test registar node, created successfully", func(t *testing.T) {
		count = 0
		request = newClientWithAccountNoNode
		c, err := NewRegistrarClient(baseURL, testMnemonic)
		require.NoError(err)

		count = 0
		request = registerNodeStatusCreated
		result, err := c.RegisterNode(node)
		require.NoError(err)
		require.Equal(nodeID, result)
	})
}

func TestUpdateNode(t *testing.T) {
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

	request = newClientWithAccountAndNode
	c, err := NewRegistrarClient(baseURL, testMnemonic)
	require.NoError(err)

	t.Run("test update node with status ok", func(t *testing.T) {
		count = 0
		request = updateNodeStatusOK
		var farmID uint64 = 2
		err = c.UpdateNode(NodeUpdate{FarmID: &farmID})
		require.NoError(err)
	})

	t.Run("test update node uptime", func(t *testing.T) {
		request = updateNodeSendUptimeReport

		report := UptimeReport{
			Uptime:    40 * 60,
			Timestamp: time.Now().Unix(),
		}
		err = c.ReportUptime(report)
		require.NoError(err)
	})
}

func TestGetNode(t *testing.T) {
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

	request = newClientWithAccountAndNode
	c, err := NewRegistrarClient(baseURL, testMnemonic)
	require.NoError(err)

	t.Run("test get node status not found", func(t *testing.T) {
		count = 0
		request = getNodeWithIDStatusNotFound
		_, err = c.GetNode(nodeID)
		require.Error(err)
	})

	t.Run("test get node, status ok", func(t *testing.T) {
		count = 0
		request = getNodeWithIDStatusOK
		result, err := c.GetNode(nodeID)
		require.NoError(err)
		require.Equal(node, result)
	})

	t.Run("test get node with twin id", func(t *testing.T) {
		count = 0
		request = getNodeWithTwinID
		result, err := c.GetNodeByTwinID(twinID)
		require.NoError(err)
		require.Equal(node, result)
	})

	t.Run("test list nodes of specific farm", func(t *testing.T) {
		count = 0
		request = listNodesInFarm
		id := farmID
		result, err := c.ListNodes(NodeFilter{FarmID: &id})
		require.NoError(err)
		require.Equal([]Node{node}, result)
	})
}

func TestGetNodeCapacityRewards(t *testing.T) {
	var request int
	var count int
	require := require.New(t)

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
	c, err := NewRegistrarClient(baseURL)
	require.NoError(err)

	t.Run("test get node capacity rewards, status ok", func(t *testing.T) {
		request = getNodeCapacityRewardsWithStatusOK
		resp, err := c.GetNodeCapacityRewards(nodeID)
		require.NoError(err)
		require.Equal(NodeCapacityReward{}, resp)
	})

	t.Run("get node rewards for non-existing node", func(t *testing.T) {
		request = getNodeCapacityRewardsWithStatusNotFound
		_, err := c.GetNodeCapacityRewards(nodeID)
		require.Error(err)
	})

	t.Run("no reports available, status UnprocessableEntity", func(t *testing.T) {
		request = getNodeCapacityRewardsWithStatusUnprocessableEntity
		res, err := c.GetNodeCapacityRewards(nodeID)
		require.Error(err)
		require.Equal(NodeCapacityReward{}, res)

	})

	t.Run("node with partial uptime rewards calculation", func(t *testing.T) {
		request = getNodeCapacityRewardsWithPartialUptime
		res, err := c.GetNodeCapacityRewards(nodeID)
		require.NoError(err)
		expected := NodeCapacityReward{
			FarmerReward:     60.0,
			TFReward:         20.0,
			FPReward:         20.0,
			Total:            100.0,
			UpTimePercentage: 75.0,
		}
		require.Equal(expected, res)
		// Verify reward distribution percentages are correct
		require.Equal(0.6, res.FarmerReward/res.Total)
		require.Equal(0.2, res.TFReward/res.Total)
		require.Equal(0.2, res.FPReward/res.Total)
	})

	t.Run("bad request due to invalid node ID format", func(t *testing.T) {
		request = getNodeCapacityRewardsWithBadRequest
		res, err := c.GetNodeCapacityRewards(nodeID)
		require.Error(err)
		require.Equal(NodeCapacityReward{}, res)
	})

	t.Run("internal server error when calculating rewards", func(t *testing.T) {
		request = getNodeCapacityRewardsWithServerError
		res, err := c.GetNodeCapacityRewards(nodeID)
		require.Error(err)
		require.Equal(NodeCapacityReward{}, res)
	})
}
