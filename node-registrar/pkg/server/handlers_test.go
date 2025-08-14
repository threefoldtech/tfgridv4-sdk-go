package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/mocks"
	"go.uber.org/mock/gomock"
)

func setupTestServer(mockDB *mocks.MockDB) Server {
	gin.SetMode(gin.TestMode)
	return NewServer(mockDB, "test", 1)
}

func createTestContext(method, url string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, w
}

func setAuthContext(c *gin.Context, twinID uint64) {
	ctx := context.WithValue(c.Request.Context(), twinIDKey{}, twinID)
	c.Request = c.Request.WithContext(ctx)
}

// Test data factories
func createTestFarm(id uint64, name string, twinID uint64) db.Farm {
	return db.Farm{
		FarmID:         id,
		FarmName:       name,
		TwinID:         twinID,
		StellarAddress: "G" + strings.Repeat("D", 55),
		Dedicated:      false,
	}
}

func createTestNode(nodeID, farmID, twinID uint64, lastSeen time.Time) db.Node {
	return db.Node{
		NodeID:       nodeID,
		FarmID:       farmID,
		TwinID:       twinID,
		LastSeen:     lastSeen,
		Online:       false,
		Approved:     true,
		SecureBoot:   true,
		Virtualized:  false,
		SerialNumber: fmt.Sprintf("SN%d", nodeID),
		Resources:    db.Resources{HRU: 1000, SRU: 500, CRU: 8, MRU: 16},
		Location:     db.Location{Country: "US", City: "NYC", Longitude: "-74.0", Latitude: "40.7"},
		Interfaces:   []db.Interface{{Name: "eth0", Mac: "00:11:22:33:44:55", IPs: []string{"192.168.1.10"}}},
	}
}

func TestListFarmsHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockDB)
		expectedStatus int
		expectedBody   func([]db.Farm) interface{}
		expectError    bool
		errorMessage   string
	}{
		{
			name: "successful farms listing",
			setupMock: func(mockDB *mocks.MockDB) {
				expectedFarms := []db.Farm{
					createTestFarm(1, "TestFarm1", 100),
					createTestFarm(2, "TestFarm2", 200),
				}
				mockDB.EXPECT().
					ListFarms(gomock.Any(), gomock.Any()).
					Return(expectedFarms, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(farms []db.Farm) interface{} {
				return farms
			},
		},
		{
			name: "empty farms list",
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					ListFarms(gomock.Any(), gomock.Any()).
					Return([]db.Farm{}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(farms []db.Farm) interface{} {
				return []db.Farm{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			server := setupTestServer(mockDB)
			c, w := createTestContext("GET", "/farms", nil)

			server.listFarmsHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			var farms []db.Farm
			err := json.Unmarshal(w.Body.Bytes(), &farms)
			assert.NoError(t, err)
			expectedBody := tt.expectedBody(farms)
			actualBodyBytes, _ := json.Marshal(farms)
			expectedBodyBytes, _ := json.Marshal(expectedBody)
			assert.JSONEq(t, string(expectedBodyBytes), string(actualBodyBytes))
		})
	}
}

func TestGetFarmHandler(t *testing.T) {
	tests := []struct {
		name           string
		farmID         string
		setupMock      func(*mocks.MockDB)
		expectedStatus int
		expectedBody   func(db.Farm) interface{}
		expectError    bool
		errorMessage   string
	}{
		{
			name:   "successful farm retrieval",
			farmID: "1",
			setupMock: func(mockDB *mocks.MockDB) {
				expectedFarm := createTestFarm(1, "TestFarm", 100)
				mockDB.EXPECT().
					GetFarm(uint64(1)).
					Return(expectedFarm, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(farm db.Farm) interface{} {
				return farm
			},
		},
		{
			name:   "farm not found",
			farmID: "999",
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					GetFarm(uint64(999)).
					Return(db.Farm{}, db.ErrRecordNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			errorMessage:   "could not find any records",
		},
		{
			name:           "invalid farm ID",
			farmID:         "0",
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "Invalid farm_id: farm_id cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			server := setupTestServer(mockDB)
			c, w := createTestContext("GET", "/farms/"+tt.farmID, nil)
			c.Params = []gin.Param{{Key: "farm_id", Value: tt.farmID}}

			server.getFarmHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			var farm db.Farm
			err := json.Unmarshal(w.Body.Bytes(), &farm)
			assert.NoError(t, err)
			expectedBody := tt.expectedBody(farm)
			actualBodyBytes, _ := json.Marshal(farm)
			expectedBodyBytes, _ := json.Marshal(expectedBody)
			assert.JSONEq(t, string(expectedBodyBytes), string(actualBodyBytes))
		})
	}
}

func TestCreateFarmHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.MockDB)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedBody   func(uint64) interface{}
		expectError    bool
		errorMessage   string
	}{
		{
			name:        "successful farm creation",
			requestBody: createTestFarm(0, "TestFarm123", 100),
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					CreateFarm(gomock.Any()).
					Return(uint64(1), nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(farmID uint64) interface{} {
				return map[string]interface{}{"farm_id": float64(farmID)}
			},
		},
		{
			name:        "farm already exists",
			requestBody: createTestFarm(0, "ExistingFarm", 100),
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					CreateFarm(gomock.Any()).
					Return(uint64(0), db.ErrRecordAlreadyExists).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
			errorMessage:   "already exists",
		},
		{
			name:           "unauthorized attempt",
			requestBody:    createTestFarm(0, "TestFarm", 100),
			setupMock:      nil,
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			errorMessage:   "not authorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			server := setupTestServer(mockDB)
			c, w := createTestContext("POST", "/farms", tt.requestBody)

			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			server.createFarmHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			expectedBody := tt.expectedBody(uint64(response["farm_id"].(float64)))
			actualBodyBytes, _ := json.Marshal(response)
			expectedBodyBytes, _ := json.Marshal(expectedBody)
			assert.JSONEq(t, string(expectedBodyBytes), string(actualBodyBytes))
		})
	}
}

func TestUpdateFarmHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		farmID         string
		requestBody    interface{}
		setupMock      func(*mocks.MockDB)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectError    bool
		errorMessage   string
	}{
		{
			name:   "successful farm update",
			farmID: "1",
			requestBody: map[string]interface{}{
				"farm_name":       "NewFarmName",
				"stellar_address": "G" + strings.Repeat("B", 55),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingFarm := createTestFarm(1, "OldFarmName", 100)
				mockDB.EXPECT().
					GetFarm(uint64(1)).
					Return(existingFarm, nil).
					Times(1)
				mockDB.EXPECT().
					UpdateFarm(uint64(1), "NewFarmName", "G"+strings.Repeat("B", 55)).
					Return(nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "no update needed same values",
			farmID: "1",
			requestBody: map[string]interface{}{
				"farm_name":       "SameFarmName",
				"stellar_address": "G" + strings.Repeat("D", 55),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingFarm := createTestFarm(1, "SameFarmName", 100)
				mockDB.EXPECT().
					GetFarm(uint64(1)).
					Return(existingFarm, nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid farm_id",
			farmID:      "invalid",
			requestBody: map[string]interface{}{},
			setupMock:   nil,
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "Invalid farm_id",
		},
		{
			name:   "farm not found",
			farmID: "999",
			requestBody: map[string]interface{}{
				"farm_name":       "TestFarm",
				"stellar_address": "G" + strings.Repeat("D", 55),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					GetFarm(uint64(999)).
					Return(db.Farm{}, db.ErrRecordNotFound).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			errorMessage:   "Farm not found",
		},
		{
			name:   "unauthorized attempt",
			farmID: "1",
			requestBody: map[string]interface{}{
				"farm_name":       "TestFarm",
				"stellar_address": "G" + strings.Repeat("D", 55),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingFarm := createTestFarm(1, "OriginalFarmName", 100)
				mockDB.EXPECT().
					GetFarm(uint64(1)).
					Return(existingFarm, nil).
					Times(1)
			},
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			errorMessage:   "not authorized",
		},
		{
			name:   "invalid stellar address format",
			farmID: "1",
			requestBody: map[string]interface{}{
				"stellar_address": "InvalidAddress",
			},
			setupMock: nil,
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "failed to parse farm info",
		},
		{
			name:   "farm name too long",
			farmID: "1",
			requestBody: map[string]interface{}{
				"farm_name": strings.Repeat("A", 50),
			},
			setupMock: nil,
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "failed to parse farm info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			server := Server{db: mockDB}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var bodyBytes []byte
			var err error
			if tt.requestBody != nil {
				bodyBytes, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, "/farms/"+tt.farmID, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			c.Params = []gin.Param{
				{Key: "farm_id", Value: tt.farmID},
			}

			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			server.updateFarmHandler(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Farm was updated successfully", response["message"])
		})
	}
}

func TestListNodesHandler(t *testing.T) {
	now := time.Now()
	onlineTime := now.Add(-30 * time.Minute)
	offlineTime := now.Add(-50 * time.Minute)

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mocks.MockDB)
		expectedStatus int
		expectedBody   func([]db.Node) interface{}
		expectError    bool
		errorMessage   string
		validateOnline func([]db.Node)
	}{
		{
			name:        "successful nodes listing",
			queryParams: "",
			setupMock: func(mockDB *mocks.MockDB) {
				expectedNodes := []db.Node{
					createTestNode(1, 1, 100, onlineTime),
					createTestNode(2, 1, 200, offlineTime),
				}
				mockDB.EXPECT().
					ListNodes(gomock.Any(), gomock.Any()).
					Return(expectedNodes, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(nodes []db.Node) interface{} {
				return nodes
			},
			validateOnline: func(nodes []db.Node) {
				assert.True(t, nodes[0].Online, "First node should be online (within cutoff)")
				assert.False(t, nodes[1].Online, "Second node should be offline (beyond cutoff)")
			},
		},
		{
			name:        "empty nodes list",
			queryParams: "",
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					ListNodes(gomock.Any(), gomock.Any()).
					Return([]db.Node{}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(nodes []db.Node) interface{} {
				return []db.Node{}
			},
		},
		{
			name:        "nodes with zero LastSeen (never seen)",
			queryParams: "",
			setupMock: func(mockDB *mocks.MockDB) {
				expectedNodes := []db.Node{
					createTestNode(3, 2, 300, time.Time{}),
				}
				mockDB.EXPECT().
					ListNodes(gomock.Any(), gomock.Any()).
					Return(expectedNodes, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(nodes []db.Node) interface{} {
				return nodes
			},
			validateOnline: func(nodes []db.Node) {
				assert.False(t, nodes[0].Online, "Node with zero LastSeen should be offline")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			server := setupTestServer(mockDB)
			c, w := createTestContext("GET", "/nodes"+tt.queryParams, nil)

			server.listNodesHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			var nodes []db.Node
			err := json.Unmarshal(w.Body.Bytes(), &nodes)
			assert.NoError(t, err)

			expectedBody := tt.expectedBody(nodes)
			actualBodyBytes, _ := json.Marshal(nodes)
			expectedBodyBytes, _ := json.Marshal(expectedBody)
			assert.JSONEq(t, string(expectedBodyBytes), string(actualBodyBytes))

			if tt.validateOnline != nil {
				tt.validateOnline(nodes)
			}
		})
	}
}

func TestGetNodeHandler(t *testing.T) {
	now := time.Now()
	onlineTime := now.Add(-30 * time.Minute)

	tests := []struct {
		name           string
		nodeID         string
		setupMock      func(*mocks.MockDB)
		expectedStatus int
		expectedBody   interface{}
		expectError    bool
		errorMessage   string
		validateOnline func(db.Node)
	}{
		{
			name:   "successful node retrieval online node",
			nodeID: "1",
			setupMock: func(mockDB *mocks.MockDB) {
				expectedNode := createTestNode(1, 1, 100, onlineTime)
				expectedNode.SerialNumber = "ABC123"
				mockDB.EXPECT().
					GetNode(uint64(1)).
					Return(expectedNode, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			validateOnline: func(node db.Node) {
				assert.True(t, node.Online, "Node should be online (within cutoff)")
				assert.Equal(t, uint64(1), node.NodeID)
				assert.Equal(t, uint64(100), node.TwinID)
				assert.Equal(t, "ABC123", node.SerialNumber)
			},
		},
		{
			name:   "node with zero LastSeen",
			nodeID: "3",
			setupMock: func(mockDB *mocks.MockDB) {
				expectedNode := createTestNode(3, 3, 300, time.Time{})
				mockDB.EXPECT().
					GetNode(uint64(3)).
					Return(expectedNode, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			validateOnline: func(node db.Node) {
				assert.False(t, node.Online, "Node with zero LastSeen should be offline")
				assert.Equal(t, uint64(3), node.NodeID)
				assert.True(t, node.LastSeen.IsZero(), "LastSeen should be zero time")
			},
		},
		{
			name:   "node not found",
			nodeID: "999",
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					GetNode(uint64(999)).
					Return(db.Node{}, db.ErrRecordNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			errorMessage:   "could not find any records",
		},
		{
			name:           "invalid node ID - zero",
			nodeID:         "0",
			setupMock:      nil,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "Invalid node id",
		},
		{
			name:   "node with recent LastSeen (exactly at cutoff)",
			nodeID: "4",
			setupMock: func(mockDB *mocks.MockDB) {
				exactCutoffTime := now.Add(-40 * time.Minute)
				expectedNode := createTestNode(4, 4, 400, exactCutoffTime)
				mockDB.EXPECT().
					GetNode(uint64(4)).
					Return(expectedNode, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			validateOnline: func(node db.Node) {
				assert.False(t, node.Online, "Node at exactly cutoff time should be offline")
				assert.Equal(t, uint64(4), node.NodeID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			server := setupTestServer(mockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			c, w := createTestContext("GET", "/nodes/"+tt.nodeID, nil)
			c.Params = []gin.Param{{Key: "node_id", Value: tt.nodeID}}

			server.getNodeHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			var node db.Node
			err := json.Unmarshal(w.Body.Bytes(), &node)
			assert.NoError(t, err)

			if tt.validateOnline != nil {
				tt.validateOnline(node)
			}
		})
	}
}

func TestRegisterNodeHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.MockDB)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectError    bool
		errorMessage   string
		expectedBody   func(uint64) interface{}
	}{
		{
			name:        "successful node registration",
			requestBody: createTestNode(0, 1, 100, time.Now()),
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					RegisterNode(gomock.Any()).
					Return(uint64(5), nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(nodeID uint64) interface{} {
				return map[string]interface{}{"node_id": float64(nodeID)}
			},
		},
		{
			name:        "node already exists",
			requestBody: createTestNode(0, 1, 100, time.Now()),
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					RegisterNode(gomock.Any()).
					Return(uint64(0), db.ErrRecordAlreadyExists).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
			errorMessage:   "already exists",
		},
		{
			name:           "unauthorized attempt",
			requestBody:    createTestNode(0, 1, 100, time.Now()),
			setupMock:      nil,
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			errorMessage:   "not authorized",
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"twin_id": 100,
			},
			setupMock:      nil,
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			server := setupTestServer(mockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			c, w := createTestContext("POST", "/nodes", tt.requestBody)
			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			server.registerNodeHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				expectedResponse := tt.expectedBody(5)
				assert.Equal(t, expectedResponse, response)
			}
		})
	}
}

func TestUpdateNodeHandler(t *testing.T) {
	tests := []struct {
		name           string
		nodeID         string
		requestBody    interface{}
		setupMock      func(*mocks.MockDB)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectError    bool
		errorMessage   string
	}{
		{
			name:   "successful node update",
			nodeID: "1",
			requestBody: map[string]interface{}{
				"farm_id":       uint64(2),
				"resources":     map[string]interface{}{"hru": 1000, "sru": 500, "cru": 8, "mru": 16},
				"location":      map[string]interface{}{"country": "US", "city": "NYC", "longitude": "-74.0", "latitude": "40.7"},
				"interfaces":    []map[string]interface{}{{"name": "eth0", "mac": "00:11:22:33:44:55", "ips": []string{"192.168.1.10"}}},
				"secure_boot":   true,
				"virtualized":   false,
				"serial_number": "UPDATED123",
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingNode := createTestNode(1, 1, 100, time.Now())
				mockDB.EXPECT().
					GetNode(uint64(1)).
					Return(existingNode, nil).
					Times(1)
				mockDB.EXPECT().
					UpdateNode(uint64(1), gomock.Any()).
					Return(nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "node not found during get",
			nodeID: "999",
			requestBody: map[string]interface{}{
				"farm_id":       uint64(1),
				"resources":     map[string]interface{}{"hru": 1000, "sru": 500, "cru": 8, "mru": 16},
				"location":      map[string]interface{}{"country": "US", "city": "NYC", "longitude": "-74.0", "latitude": "40.7"},
				"interfaces":    []map[string]interface{}{{"name": "eth0", "mac": "00:11:22:33:44:55", "ips": []string{"192.168.1.10"}}},
				"secure_boot":   true,
				"virtualized":   false,
				"serial_number": "TEST123",
			},
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					GetNode(uint64(999)).
					Return(db.Node{}, db.ErrRecordNotFound).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			errorMessage:   "node not found",
		},
		{
			name:   "unauthorized update attempt",
			nodeID: "1",
			requestBody: map[string]interface{}{
				"farm_id":       uint64(1),
				"resources":     map[string]interface{}{"hru": 1000, "sru": 500, "cru": 8, "mru": 16},
				"location":      map[string]interface{}{"country": "US", "city": "NYC", "longitude": "-74.0", "latitude": "40.7"},
				"interfaces":    []map[string]interface{}{{"name": "eth0", "mac": "00:11:22:33:44:55", "ips": []string{"192.168.1.10"}}},
				"secure_boot":   true,
				"virtualized":   false,
				"serial_number": "TEST123",
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingNode := createTestNode(1, 1, 100, time.Now())
				mockDB.EXPECT().
					GetNode(uint64(1)).
					Return(existingNode, nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 200) // Different twin ID
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			errorMessage:   "not authorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			server := setupTestServer(mockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			c, w := createTestContext("PUT", "/nodes/"+tt.nodeID, tt.requestBody)
			c.Params = []gin.Param{{Key: "node_id", Value: tt.nodeID}}

			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			server.updateNodeHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "node updated successfully", response["message"])
		})
	}
}

func TestUptimeReportHandler(t *testing.T) {
	tests := []struct {
		name           string
		nodeID         string
		requestBody    interface{}
		setupMock      func(*mocks.MockDB)
		setupContext   func(*gin.Context)
		expectedStatus int
		expectError    bool
		errorMessage   string
	}{
		{
			name:   "successful uptime report",
			nodeID: "1",
			requestBody: map[string]interface{}{
				"uptime":    uint64(3600),
				"timestamp": time.Now().Unix(),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingNode := createTestNode(1, 1, 100, time.Now())
				mockDB.EXPECT().
					GetNode(uint64(1)).
					Return(existingNode, nil).
					Times(1)
				mockDB.EXPECT().
					CreateUptimeReport(gomock.Any()).
					Return(nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:   "node not found",
			nodeID: "999",
			requestBody: map[string]interface{}{
				"uptime":    uint64(3600),
				"timestamp": time.Now().Unix(),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					GetNode(uint64(999)).
					Return(db.Node{}, db.ErrRecordNotFound).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
			errorMessage:   "node not found",
		},
		{
			name:   "unauthorized attempt",
			nodeID: "1",
			requestBody: map[string]interface{}{
				"uptime":    uint64(3600),
				"timestamp": time.Now().Unix(),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingNode := createTestNode(1, 1, 100, time.Now())
				mockDB.EXPECT().
					GetNode(uint64(1)).
					Return(existingNode, nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 200) 
			},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			errorMessage:   "not authorized",
		},
		{
			name:   "invalid timestamp",
			nodeID: "1",
			requestBody: map[string]interface{}{
				"uptime":    uint64(3600),
				"timestamp": time.Now().Add(5 * time.Minute).Unix(), // Future timestamp
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingNode := createTestNode(1, 1, 100, time.Now())
				mockDB.EXPECT().
					GetNode(uint64(1)).
					Return(existingNode, nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				setAuthContext(c, 100)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "invalid timestamp hint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			server := setupTestServer(mockDB)

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			c, w := createTestContext("POST", "/nodes/"+tt.nodeID+"/uptime", tt.requestBody)
			c.Params = []gin.Param{{Key: "node_id", Value: tt.nodeID}}

			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			server.uptimeReportHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectError {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.errorMessage)
				return
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "uptime reported successfully", response["message"])
		})
	}
}
