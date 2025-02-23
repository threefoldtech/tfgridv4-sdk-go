package client

import (
	"crypto/ed25519"
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
	node      = Node{FarmID: farmID, TwinID: twinID}
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
	updateNodeStatusUnauthorized
	updateNodeSendUptimeReport
	getNodeWithIDStatusOK
	getNodeWithIDStatusNotFound
	getNodeWithTwinID
	listNodesInFarm

	farmID = 1
	nodeID = 1
	twinID = 1
)

func accountHandler(r *http.Request, request, count int, require *require.Assertions) (statusCode int, body []byte) {
	switch request {
	// NewRegistrarClient handlers
	case newClientWithAccountNoNode:
		switch count {
		case 0:
			require.Equal(r.URL.Path, "/v1/accounts")
			require.Equal(r.URL.Query().Get("public_key"), account.PublicKey)
			require.Equal(r.Method, http.MethodGet)
			resp, err := json.Marshal(account)
			require.NoError(err)
			return http.StatusOK, resp
		case 1:
			require.Equal(r.URL.Path, "/v1/nodes")
			require.Equal(r.URL.Query().Get("twin_id"), fmt.Sprint(account.TwinID))
			require.Equal(r.Method, http.MethodGet)
			return http.StatusNotFound, nil
		}

	case newClientWithAccountAndNode:
		switch count {
		case 0:
			require.Equal(r.URL.Path, "/v1/accounts")
			require.Equal(r.URL.Query().Get("public_key"), account.PublicKey)
			require.Equal(r.Method, http.MethodGet)
			resp, err := json.Marshal(account)
			require.NoError(err)
			return http.StatusOK, resp
		case 1:
			require.Equal(r.URL.Path, "/v1/nodes")
			require.Equal(r.URL.Query().Get("twin_id"), fmt.Sprint(account.TwinID))
			require.Equal(r.Method, http.MethodGet)
			resp, err := json.Marshal([]Node{{NodeID: nodeID, TwinID: account.TwinID}})
			require.NoError(err)
			return http.StatusOK, resp
		}

	case newClientWithNoAccount,

		// Accounts routes handlers
		getAccountWithPKStatusNotFount,
		updateAccountWithNoAccount:
		require.Equal(r.URL.Path, "/v1/accounts")
		require.Equal(r.URL.Query().Get("public_key"), account.PublicKey)
		require.Equal(r.Method, http.MethodGet)
		return http.StatusNotFound, nil

	case createAccountStatusCreated:
		require.Equal(r.URL.Path, "/v1/accounts")
		require.Equal(r.Method, http.MethodPost)
		require.NotEmpty(r.Body)
		resp, err := json.Marshal(account)
		require.NoError(err)
		return http.StatusCreated, resp

	case getAccountWithPKStatusOK:
		require.Equal(r.URL.Path, "/v1/accounts")
		require.Equal(r.URL.Query().Get("public_key"), account.PublicKey)
		require.Equal(r.Method, http.MethodGet)
		resp, err := json.Marshal(account)
		require.NoError(err)
		return http.StatusOK, resp

	case getAccountWithIDStatusNotFount:
		require.Equal(r.URL.Path, "/v1/accounts")
		require.Equal(r.URL.Query().Get("twin_id"), fmt.Sprint(account.TwinID))
		require.Equal(r.Method, http.MethodGet)
		return http.StatusNotFound, nil

	case getAccountWithIDStatusOK:
		require.Equal(r.URL.Path, "/v1/accounts")
		require.Equal(r.URL.Query().Get("twin_id"), fmt.Sprint(account.TwinID))
		require.Equal(r.Method, http.MethodGet)
		resp, err := json.Marshal(account)
		require.NoError(err)
		return http.StatusOK, resp

	case updateAccountWithStatusOK:
		require.Equal(r.URL.Path, "/v1/accounts/1")
		require.Equal(r.Method, http.MethodPatch)
		return http.StatusOK, nil

		// Farm routes handlers
	case createFarmStatusCreated:
		require.Equal(r.URL.Path, "/v1/farms")
		require.Equal(r.Method, http.MethodPost)
		require.NotEmpty(r.Body)
		resp, err := json.Marshal(`{"farm_id": 1}`)
		require.NoError(err)
		return http.StatusCreated, resp

	case updateFarmWithStatusOK:
		require.Equal(r.URL.Path, "/v1/farms/1")
		require.Equal(r.Method, http.MethodPatch)
		require.NotEmpty(r.Body)
		return http.StatusOK, nil

	case updateFarmWithStatusUnauthorized:
		require.Equal(r.URL.Path, "/v1/farms/1")
		require.Equal(r.Method, http.MethodPatch)
		require.NotEmpty(r.Body)
		return http.StatusUnauthorized, nil

	case getFarmWithStatusOK:
		require.Equal(r.URL.Path, "/v1/farms/1")
		require.Equal(r.Method, http.MethodGet)
		resp, err := json.Marshal(farm)
		require.NoError(err)
		return http.StatusOK, resp

	case getFarmWithStatusNotfound:
		require.Equal(r.URL.Path, "/v1/farms/1")
		require.Equal(r.Method, http.MethodGet)
		return http.StatusNotFound, nil

	}

	return http.StatusNotAcceptable, nil
}

func aliceKeys() (pk, seed []byte, err error) {
	seed, err = hex.DecodeString(aliceSeed)
	if err != nil {
		return
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	pk, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return pk, seed, fmt.Errorf("failed to get public key of provided private key")
	}

	return pk, seed, nil
}
