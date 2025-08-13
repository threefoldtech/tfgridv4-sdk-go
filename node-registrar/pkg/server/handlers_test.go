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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/mocks"
	"go.uber.org/mock/gomock"
)

func TestListFarmsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
					{
						FarmID:         1,
						FarmName:       "TestFarm1",
						TwinID:         100,
						StellarAddress: "G" + strings.Repeat("D", 55),
						Dedicated:      false,
					},
					{
						FarmID:         2,
						FarmName:       "TestFarm2",
						TwinID:         200,
						StellarAddress: "G" + strings.Repeat("E", 55),
						Dedicated:      true,
					},
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
			name: "database error",
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					ListFarms(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("database connection error")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			errorMessage:   "database connection error",
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

			server := NewServer(mockDB, "test", 1)

			req, err := http.NewRequest("GET", "/farms", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

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
			err = json.Unmarshal(w.Body.Bytes(), &farms)
			assert.NoError(t, err)
			expectedBody := tt.expectedBody(farms)
			actualBodyBytes, _ := json.Marshal(farms)
			expectedBodyBytes, _ := json.Marshal(expectedBody)
			assert.JSONEq(t, string(expectedBodyBytes), string(actualBodyBytes))
		})
	}
}

func TestGetFarmHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
				expectedFarm := db.Farm{
					FarmID:         1,
					FarmName:       "TestFarm",
					TwinID:         100,
					StellarAddress: "G" + strings.Repeat("D", 55),
					Dedicated:      false,
				}
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

			server := NewServer(mockDB, "test", 1)

			req, err := http.NewRequest("GET", "/farms/"+tt.farmID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			c.Params = []gin.Param{
				{Key: "farm_id", Value: tt.farmID},
			}

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
			err = json.Unmarshal(w.Body.Bytes(), &farm)
			assert.NoError(t, err)
			expectedBody := tt.expectedBody(farm)
			actualBodyBytes, _ := json.Marshal(farm)
			expectedBodyBytes, _ := json.Marshal(expectedBody)
			assert.JSONEq(t, string(expectedBodyBytes), string(actualBodyBytes))
		})
	}
}

func TestCreateFarmHandler(t *testing.T) {

	gin.SetMode(gin.TestMode)

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
			name: "successful farm creation",
			requestBody: db.Farm{
				FarmName:       "TestFarm123",
				TwinID:         100,
				StellarAddress: "G" + strings.Repeat("D", 55),
				Dedicated:      false,
			},
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					CreateFarm(gomock.Any()).
					Return(uint64(1), nil).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(farmID uint64) interface{} {
				return map[string]interface{}{"farm_id": float64(farmID)}
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: "invalid json",
			setupMock:   nil,
			setupContext: func(c *gin.Context) {
				c.Set("twin_id", uint64(100))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "failed to parse farm info",
		},
		{
			name: "farm already exists",
			requestBody: db.Farm{
				FarmName:       "ExistingFarm",
				TwinID:         100,
				StellarAddress: "G" + strings.Repeat("D", 55),
				Dedicated:      false,
			},
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					CreateFarm(gomock.Any()).
					Return(uint64(0), db.ErrRecordAlreadyExists).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
			errorMessage:   "already exists",
		},
		{
			name: "database error",
			requestBody: db.Farm{
				FarmName:       "TestFarm",
				TwinID:         100,
				StellarAddress: "G" + strings.Repeat("D", 55),
				Dedicated:      false,
			},
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					CreateFarm(gomock.Any()).
					Return(uint64(0), fmt.Errorf("database connection error")).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			errorMessage:   "database connection error",
		},
		{
			name: "unauthorized no twin_id in context",
			requestBody: db.Farm{
				FarmName:       "TestFarm",
				TwinID:         100,
				StellarAddress: "G" + strings.Repeat("D", 55),
				Dedicated:      false,
			},
			setupMock:      nil,
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			errorMessage:   "not authorized",
		},
		{
			name: "unauthorized mismatched twin_id",
			requestBody: db.Farm{
				FarmName:       "TestFarm",
				TwinID:         100,
				StellarAddress: "G" + strings.Repeat("D", 55),
				Dedicated:      false,
			},
			setupMock: nil,
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(200))
				c.Request = c.Request.WithContext(ctx)
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

			if tt.setupMock != nil {
				tt.setupMock(mockDB)
			}

			server := NewServer(mockDB, "test", 1)

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest("POST", "/farms", bytes.NewBuffer(bodyBytes))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

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
			err = json.Unmarshal(w.Body.Bytes(), &response)
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
				existingFarm := db.Farm{
					FarmID:         1,
					FarmName:       "OldFarmName",
					TwinID:         100,
					StellarAddress: "G" + strings.Repeat("D", 55),
					Dedicated:      false,
				}
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
				existingFarm := db.Farm{
					FarmID:         1,
					FarmName:       "SameFarmName",
					TwinID:         100,
					StellarAddress: "G" + strings.Repeat("D", 55),
					Dedicated:      false,
				}
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
			name:        "invalid JSON body",
			farmID:      "1",
			requestBody: "invalid json",
			setupMock:   nil,
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "failed to parse farm info",
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
			name:   "database error",
			farmID: "1",
			requestBody: map[string]interface{}{
				"farm_name":       "TestFarm",
				"stellar_address": "G" + strings.Repeat("D", 55),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				mockDB.EXPECT().
					GetFarm(uint64(1)).
					Return(db.Farm{}, fmt.Errorf("database connection error")).
					Times(1)
			},
			setupContext: func(c *gin.Context) {
				ctx := context.WithValue(c.Request.Context(), twinIDKey{}, uint64(100))
				c.Request = c.Request.WithContext(ctx)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			errorMessage:   "Database error",
		},
		{
			name:   "unauthorized no twin_id in context",
			farmID: "1",
			requestBody: map[string]interface{}{
				"farm_name":       "TestFarm",
				"stellar_address": "G" + strings.Repeat("D", 55),
			},
			setupMock: func(mockDB *mocks.MockDB) {
				existingFarm := db.Farm{
					FarmID:         1,
					FarmName:       "OriginalFarmName",
					TwinID:         100,
					StellarAddress: "G" + strings.Repeat("D", 55),
					Dedicated:      false,
				}
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
