package client

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
)

var (
	aliceSeed = "e5be9a5092b81bca64be81d212e7f2f9eba183bb7a90954f7b76361f6edb5c0a"
	account   = Account{TwinID: 1, Relays: []string{}, RMBEncKey: ""}
	farm      = Farm{FarmID: 1, FarmName: "freeFarm", TwinID: 1}
	node      = Node{NodeID: 1, FarmID: farmID, TwinID: twinID}
)

const (
	newClientWithNoAccount = iota
	newClientWithAccountNoNode
	newClientWithAccountAndNode

	createAccountStatusCreated
	updateAccountWithNoAccount
	updateAccountWithStatusOK
	getAccountWithPKStatusOK
	getAccountWithPKStatusNotFount
	getAccountWithIDStatusOK
	getAccountWithIDStatusNotFount

	createFarmStatusCreated
	createFarmStatusConflict
	updateFarmWithStatusOK
	updateFarmWithStatusUnauthorized
	getFarmWithStatusNotfound
	getFarmWithStatusOK

	registerNodeStatusCreated
	registerNodeWithNoAccount
	registerNodeStatusConflict
	updateNodeStatusOK
	updateNodeSendUptimeReport
	getNodeWithIDStatusOK
	getNodeWithIDStatusNotFound
	getNodeWithTwinID
	listNodesInFarm

	farmID uint64 = 1
	nodeID uint64 = 1
	twinID uint64 = 1
)

func serverHandler(r *http.Request, request, count int, require *require.Assertions) (statusCode int, body []byte) {
	switch request {
	// NewRegistrarClient handlers
	case newClientWithAccountNoNode:
		switch count {
		case 0:
			require.Equal("/v1/accounts", r.URL.Path)
			require.Equal(account.PublicKey, r.URL.Query().Get("public_key"))
			require.Equal(http.MethodGet, r.Method)
			resp, err := json.Marshal(account)
			require.NoError(err)
			return http.StatusOK, resp
		case 1:
			require.Equal("/v1/nodes", r.URL.Path)
			require.Equal(fmt.Sprint(account.TwinID), r.URL.Query().Get("twin_id"))
			require.Equal(http.MethodGet, r.Method)
			return http.StatusNotFound, nil
		}

	case newClientWithAccountAndNode:
		switch count {
		case 0:
			require.Equal("/v1/accounts", r.URL.Path)
			require.Equal(account.PublicKey, r.URL.Query().Get("public_key"))
			require.Equal(http.MethodGet, r.Method)
			resp, err := json.Marshal(account)
			require.NoError(err)
			return http.StatusOK, resp
		case 1:
			require.Equal("/v1/nodes", r.URL.Path)
			require.Equal(fmt.Sprint(account.TwinID), r.URL.Query().Get("twin_id"))
			require.Equal(http.MethodGet, r.Method)
			resp, err := json.Marshal([]Node{node})
			require.NoError(err)
			return http.StatusOK, resp
		}

		// Accounts routes handlers
	case createAccountStatusCreated:
		require.Equal("/v1/accounts", r.URL.Path)
		require.Equal(http.MethodPost, r.Method)
		require.NotEmpty(r.Body)
		resp, err := json.Marshal(account)
		require.NoError(err)
		return http.StatusCreated, resp

	case getAccountWithPKStatusOK:
		require.Equal("/v1/accounts", r.URL.Path)
		require.Equal(account.PublicKey, r.URL.Query().Get("public_key"))
		require.Equal(http.MethodGet, r.Method)
		resp, err := json.Marshal(account)
		require.NoError(err)
		return http.StatusOK, resp

	case getAccountWithIDStatusNotFount:
		require.Equal("/v1/accounts", r.URL.Path)
		require.Equal(fmt.Sprint(account.TwinID), r.URL.Query().Get("twin_id"))
		require.Equal(http.MethodGet, r.Method)
		return http.StatusNotFound, nil

	case getAccountWithIDStatusOK:
		require.Equal("/v1/accounts", r.URL.Path)
		require.Equal(fmt.Sprint(account.TwinID), r.URL.Query().Get("twin_id"))
		require.Equal(http.MethodGet, r.Method)
		resp, err := json.Marshal(account)
		require.NoError(err)
		return http.StatusOK, resp

	case updateAccountWithStatusOK:
		require.Equal("/v1/accounts/1", r.URL.Path)
		require.Equal(http.MethodPatch, r.Method)
		return http.StatusOK, nil

		// Farm routes handlers
	case createFarmStatusConflict:
		require.Equal("/v1/farms", r.URL.Path)
		require.Equal(http.MethodPost, r.Method)
		require.NotEmpty(r.Body)
		return http.StatusConflict, nil

	case createFarmStatusCreated:
		require.Equal("/v1/farms", r.URL.Path)
		require.Equal(http.MethodPost, r.Method)
		require.NotEmpty(r.Body)
		resp, err := json.Marshal(map[string]uint64{"farm_id": farmID})
		require.NoError(err)
		return http.StatusCreated, resp

	case updateFarmWithStatusOK:
		require.Equal("/v1/farms/1", r.URL.Path)
		require.Equal(http.MethodPatch, r.Method)
		require.NotEmpty(r.Body)
		return http.StatusOK, nil

	case getFarmWithStatusOK:
		require.Equal("/v1/farms/1", r.URL.Path)
		require.Equal(http.MethodGet, r.Method)
		resp, err := json.Marshal(farm)
		require.NoError(err)
		return http.StatusOK, resp

	case getFarmWithStatusNotfound:
		require.Equal("/v1/farms/1", r.URL.Path)
		require.Equal(http.MethodGet, r.Method)
		return http.StatusNotFound, nil

		// Node routes handlers
	case registerNodeStatusConflict:
		require.Equal("/v1/nodes", r.URL.Path)
		require.Equal(http.MethodPost, r.Method)
		return http.StatusConflict, nil

	case registerNodeStatusCreated:
		require.Equal("/v1/nodes", r.URL.Path)
		require.Equal(http.MethodPost, r.Method)
		require.NotEmpty(r.Body)
		resp, err := json.Marshal(map[string]uint64{"node_id": nodeID})
		require.NoError(err)
		return http.StatusCreated, resp

	case updateNodeStatusOK:
		switch count {
		case 0:
			require.Equal("/v1/nodes/1", r.URL.Path)
			require.Equal(http.MethodGet, r.Method)
			resp, err := json.Marshal(node)
			require.NoError(err)
			return http.StatusOK, resp
		case 1:
			require.Equal("/v1/nodes/1", r.URL.Path)
			require.Equal(http.MethodPatch, r.Method)
			require.NotEmpty(r.Body)
			return http.StatusOK, nil
		}

	case updateNodeSendUptimeReport:
		require.Equal("/v1/nodes/1", r.URL.Path)
		require.Equal(http.MethodPatch, r.Method)
		require.NotEmpty(r.Body)
		return http.StatusOK, nil

	case getNodeWithIDStatusOK:
		require.Equal("/v1/nodes/1", r.URL.Path)
		require.Equal(http.MethodGet, r.Method)
		resp, err := json.Marshal(node)
		require.NoError(err)
		return http.StatusOK, resp

	case getNodeWithIDStatusNotFound:
		require.Equal("/v1/nodes/1", r.URL.Path)
		require.Equal(http.MethodGet, r.Method)
		return http.StatusNotFound, nil

	case getNodeWithTwinID:
		require.Equal("/v1/nodes", r.URL.Path)
		require.Equal(fmt.Sprint(account.TwinID), r.URL.Query().Get("twin_id"))
		require.Equal(http.MethodGet, r.Method)
		resp, err := json.Marshal([]Node{node})
		require.NoError(err)
		return http.StatusOK, resp

	case listNodesInFarm:
		require.Equal("/v1/nodes", r.URL.Path)
		require.Equal(fmt.Sprint(farmID), r.URL.Query().Get("farm_id"))
		require.Equal(http.MethodGet, r.Method)
		resp, err := json.Marshal([]Node{node})
		require.NoError(err)
		return http.StatusOK, resp

	// unauthorized requests
	case newClientWithNoAccount,
		getAccountWithPKStatusNotFount,
		updateAccountWithNoAccount,
		updateFarmWithStatusUnauthorized,
		registerNodeWithNoAccount:
		require.Equal("/v1/accounts", r.URL.Path)
		require.Equal(account.PublicKey, r.URL.Query().Get("public_key"))
		require.Equal(http.MethodGet, r.Method)
		return http.StatusNotFound, nil

	}

	return http.StatusNotAcceptable, nil
}

func aliceKeys() (pk, seed []byte, pkBase64 string, err error) {
	seed, err = hex.DecodeString(aliceSeed)
	if err != nil {
		return
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	pk, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return pk, seed, pkBase64, fmt.Errorf("failed to get public key of provided private key")
	}

	pkBase64 = base64.StdEncoding.EncodeToString(pk)

	return pk, seed, pkBase64, nil
}
