package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
)

func setupTestServer() *Server {
	gin.SetMode(gin.TestMode)

	server := &Server{
		router:        gin.New(),
		db:            db.Database{},
		network:       "test",
		adminTwinID:   1,
		csrfSecret:    "test-csrf-secret-32-characters-long",
		sessionSecret: "test-session-secret-32-characters-long",
	}

	server.SetupRoutes()
	return server
}

func TestCSRFTokenEndpoint(t *testing.T) {
	server := setupTestServer()

	req, err := http.NewRequest("GET", "/csrf-token", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "csrf_token")
	assert.NotEmpty(t, response["csrf_token"])
}

func TestCSRFProtectionOnPOSTRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	sessionSecret := "test-session-secret-32-characters-long"
	csrfSecret := "test-csrf-secret-32-characters-long"

	server := &Server{
		router:        router,
		csrfSecret:    csrfSecret,
		sessionSecret: sessionSecret,
	}

	router.Use(server.SessionMiddleware(sessionSecret))
	router.Use(server.CSRFMiddleware(csrfSecret))

	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testData := map[string]interface{}{
		"test": "data",
	}

	jsonData, err := json.Marshal(testData)
	require.NoError(t, err)

	t.Run("POST without CSRF token should fail", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "CSRF token validation failed", response["error"])
		assert.Equal(t, "CSRF_TOKEN_INVALID", response["code"])
	})

	t.Run("POST with invalid CSRF token should fail", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", "invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestCSRFProtectionOnGETRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	sessionSecret := "test-session-secret-32-characters-long"
	csrfSecret := "test-csrf-secret-32-characters-long"

	server := &Server{
		router:        router,
		csrfSecret:    csrfSecret,
		sessionSecret: sessionSecret,
	}

	router.Use(server.SessionMiddleware(sessionSecret))
	router.Use(server.CSRFMiddleware(csrfSecret))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("GET requests should not require CSRF token", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSessionMiddleware(t *testing.T) {
	server := setupTestServer()

	req, err := http.NewRequest("GET", "/csrf-token", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// Check that session cookie is set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if strings.Contains(cookie.Name, "session") {
			sessionCookie = cookie
			break
		}
	}

	assert.NotNil(t, sessionCookie, "Session cookie should be set")
	if sessionCookie != nil {
		assert.True(t, sessionCookie.HttpOnly, "Session cookie should be HttpOnly")
		assert.Equal(t, http.SameSiteLaxMode, sessionCookie.SameSite, "Session cookie should use SameSite=Lax")
	}
}

func TestCSRFMiddlewareConfiguration(t *testing.T) {
	server := setupTestServer()

	// Test that CSRF middleware is properly configured
	assert.NotEmpty(t, server.csrfSecret)
	assert.NotEmpty(t, server.sessionSecret)
	assert.GreaterOrEqual(t, len(server.csrfSecret), 32)
	assert.GreaterOrEqual(t, len(server.sessionSecret), 32)
}

func TestCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Auth", "X-CSRF-Token"},
		ExposeHeaders:    []string{"X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add a test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, err := http.NewRequest("OPTIONS", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "X-CSRF-Token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	assert.Contains(t, allowHeaders, "X-Csrf-Token", "Allow headers should contain X-Csrf-Token: %s", allowHeaders)

	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}
