package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/mocks"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
	"go.uber.org/mock/gomock"
)

// Test AuthMiddleware with ED25519
func TestAuthMiddleware_ED25519(t *testing.T) {
	gin.SetMode(gin.TestMode)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	validTwinID := uint64(123)
	validAccount := db.Account{
		TwinID:    validTwinID,
		PublicKey: base64.StdEncoding.EncodeToString(publicKey),
		Relays:    []string{"relay1"},
		RMBEncKey: "test_key",
	}

	createValidAuth := func(twinID uint64, timeOffset time.Duration) (string, string) {
		timestamp := time.Now().Add(timeOffset).Unix()
		challenge := fmt.Sprintf("%d:%d", timestamp, twinID)
		challengeB64 := base64.StdEncoding.EncodeToString([]byte(challenge))
		signature := ed25519.Sign(privateKey, []byte(challenge))
		signatureB64 := base64.StdEncoding.EncodeToString(signature)
		return challengeB64, signatureB64
	}

	t.Run("expired challenge", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockDB(ctrl)
		server := Server{db: mockDB}

		router := gin.New()
		router.Use(server.AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		challengeB64, signatureB64 := createValidAuth(validAccount.TwinID, -time.Hour)
		authHeader := challengeB64 + ":" + signatureB64

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(AuthHeader, authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Expired challenge")
	})
	t.Run("signature verification failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockDB(ctrl)
		mockDB.EXPECT().
			GetAccount(validAccount.TwinID).
			Return(validAccount, nil).
			Times(1)

		server := Server{db: mockDB}

		router := gin.New()
		router.Use(server.AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		timestamp := time.Now().Unix()
		challenge := fmt.Sprintf("%d:%d", timestamp, validAccount.TwinID)
		challengeB64 := base64.StdEncoding.EncodeToString([]byte(challenge))

		wrongSignature := ed25519.Sign(privateKey, []byte("wrong_challenge"))
		signatureB64 := base64.StdEncoding.EncodeToString(wrongSignature)
		authHeader := challengeB64 + ":" + signatureB64

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(AuthHeader, authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Signature verification failed")
	})
	t.Run("successful authentication", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockDB(ctrl)
		mockDB.EXPECT().
			GetAccount(validAccount.TwinID).
			Return(validAccount, nil).
			Times(1)

		server := Server{db: mockDB}

		router := gin.New()
		nextCalled := false

		router.Use(server.AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			nextCalled = true
			twinIDValue := c.Request.Context().Value(twinIDKey{})
			assert.NotNil(t, twinIDValue)
			assert.Equal(t, validAccount.TwinID, twinIDValue.(uint64))
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		challengeB64, signatureB64 := createValidAuth(validAccount.TwinID, 0)
		authHeader := challengeB64 + ":" + signatureB64

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(AuthHeader, authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextCalled)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})
}

// Test AuthMiddleware with SR25519
func TestAuthMiddleware_SR25519(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := sr25519.Scheme{}
	keyPair, err := scheme.Generate()
	assert.NoError(t, err)

	publicKey := keyPair.Public()
	validTwinID := uint64(124)
	validAccount := db.Account{
		TwinID:    validTwinID,
		PublicKey: base64.StdEncoding.EncodeToString(publicKey),
		Relays:    []string{"relay1"},
		RMBEncKey: "test_key",
	}

	createValidAuth := func(twinID uint64, timeOffset time.Duration) (string, string) {
		timestamp := time.Now().Add(timeOffset).Unix()
		challenge := fmt.Sprintf("%d:%d", timestamp, twinID)
		challengeB64 := base64.StdEncoding.EncodeToString([]byte(challenge))
		signature, err := keyPair.Sign([]byte(challenge))
		assert.NoError(t, err)
		signatureB64 := base64.StdEncoding.EncodeToString(signature)
		return challengeB64, signatureB64
	}
	t.Run("expired challenge", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockDB(ctrl)
		server := Server{db: mockDB}

		router := gin.New()
		router.Use(server.AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		challengeB64, signatureB64 := createValidAuth(validAccount.TwinID, -time.Hour)
		authHeader := challengeB64 + ":" + signatureB64

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(AuthHeader, authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Expired challenge")
	})
	t.Run("signature verification failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockDB(ctrl)
		mockDB.EXPECT().
			GetAccount(validAccount.TwinID).
			Return(validAccount, nil).
			Times(1)

		server := Server{db: mockDB}

		router := gin.New()
		router.Use(server.AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		timestamp := time.Now().Unix()
		challenge := fmt.Sprintf("%d:%d", timestamp, validAccount.TwinID)
		challengeB64 := base64.StdEncoding.EncodeToString([]byte(challenge))

		wrongSignature, err := keyPair.Sign([]byte("wrong_challenge"))
		assert.NoError(t, err)
		signatureB64 := base64.StdEncoding.EncodeToString(wrongSignature)
		authHeader := challengeB64 + ":" + signatureB64

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(AuthHeader, authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Signature verification failed")
	})
	t.Run("successful authentication", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := mocks.NewMockDB(ctrl)
		mockDB.EXPECT().
			GetAccount(validAccount.TwinID).
			Return(validAccount, nil).
			Times(1)

		server := Server{db: mockDB}

		router := gin.New()
		nextCalled := false

		router.Use(server.AuthMiddleware())
		router.GET("/test", func(c *gin.Context) {
			nextCalled = true
			twinIDValue := c.Request.Context().Value(twinIDKey{})
			assert.NotNil(t, twinIDValue)
			assert.Equal(t, validAccount.TwinID, twinIDValue.(uint64))
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		challengeB64, signatureB64 := createValidAuth(validAccount.TwinID, 0)
		authHeader := challengeB64 + ":" + signatureB64

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(AuthHeader, authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextCalled)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})
}
