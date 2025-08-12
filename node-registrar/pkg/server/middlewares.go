package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
	csrf "github.com/utrack/gin-csrf"
)

// twinKeyID is where the twin key is stored
type twinIDKey struct{}

const (
	AuthHeader        = "X-Auth"
	ChallengeValidity = 1 * time.Minute
)

// AuthMiddleware is a middleware function that authenticates incoming requests based on the X-Auth header.
// It verifies the challenge and signature provided in the header against the account's public key stored in the database.
// If the authentication fails, it aborts the request with an appropriate error status and message.
// When authentication succeeds, it set the twinID in the contexct so handlers can trust that the requester is the owner of that Account/Twin.
// Authorization must be checked next independently by handlers or other middlewares.
// header format `Challenge:Signature`
// - chalange format: base64(message) where the message is `timestampStr:twinIDStr`
// - signature format: base64(ed25519_or_sr22519_signature)
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract and validate headers
		authHeader := c.GetHeader(AuthHeader)
		if authHeader == "" {
			abortWithError(c, http.StatusUnauthorized, "Authorization header required")
			return
		}

		parts := strings.Split(authHeader, ":")
		if len(parts) != 2 {
			abortWithError(c, http.StatusBadRequest, "Invalid header format. Use 'Challenge:Signature'")
			return
		}

		challengeB64, signatureB64 := parts[0], parts[1]

		// Decode and validate challenge
		challenge, err := base64.StdEncoding.DecodeString(challengeB64)
		if err != nil {
			log.Debug().Err(err).Msg("failed to deconde challenge")
			abortWithError(c, http.StatusBadRequest, "Invalid challenge encoding")
			return
		}

		challengeParts := strings.Split(string(challenge), ":")
		if len(challengeParts) != 2 {
			abortWithError(c, http.StatusBadRequest, "Invalid challenge format")
			return
		}

		timestampStr, twinIDStr := challengeParts[0], challengeParts[1]

		// Validate timestamp
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			log.Debug().Err(err).Msg("invalid timestamp")
			abortWithError(c, http.StatusBadRequest, "Invalid timestamp")
			return
		}

		if time.Since(time.Unix(timestamp, 0)) > ChallengeValidity {
			abortWithError(c, http.StatusUnauthorized, "Expired challenge")
			return
		}

		twinID, err := strconv.ParseUint(twinIDStr, 10, 64)
		if err != nil {
			log.Debug().Err(err).Msg("invalid twin id format")
			abortWithError(c, http.StatusBadRequest, "Invalid twin ID format")
			return
		}

		account, err := s.db.GetAccount(twinID)
		if err != nil {
			log.Debug().Err(err).Uint64("twinID", twinID).Msg("failed to get account")
			handleDatabaseError(c, err)
			return
		}

		storedPK, err := base64.StdEncoding.DecodeString(account.PublicKey)
		if err != nil {
			log.Debug().Err(err).Msg("failed to get invalid stored public key")
			abortWithError(c, http.StatusBadRequest, fmt.Sprintf("invalid stored public key: %v", err))
			return
		}

		sig, err := base64.StdEncoding.DecodeString(signatureB64)
		if err != nil {
			log.Debug().Err(err).Msg("invalid signature encoding")
			abortWithError(c, http.StatusBadRequest, "Invalid signature encoding")
			return
		}

		// Verify signature (supports both ED25519 and SR25519)
		if err := verifySignature(storedPK, challenge, sig); err != nil {
			log.Debug().Err(err).Msg("signature verification failed")
			abortWithError(c, http.StatusUnauthorized, fmt.Sprintf("Signature verification failed: %v", err))
			return
		}

		// Store verified twin ID in context, must be checked form the handlers to ensure altred resources belongs to same user
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), twinIDKey{}, twinID))
		c.Next()
	}
}

// CSRFMiddleware creates and configures CSRF protection middleware
// It requires sessions to be configured before this middleware
func (s *Server) CSRFMiddleware(secret string) gin.HandlerFunc {
	return csrf.Middleware(csrf.Options{
		Secret: secret,
		ErrorFunc: func(c *gin.Context) {
			log.Warn().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Str("remote_addr", c.ClientIP()).
				Msg("CSRF token validation failed")

			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
				"code":  "CSRF_TOKEN_INVALID",
			})
			c.Abort()
		},
		// Only check CSRF for state-changing methods
		IgnoreMethods: []string{"GET", "HEAD", "OPTIONS"},
	})
}

// SessionMiddleware configures session management required for CSRF protection
func (s *Server) SessionMiddleware(sessionSecret string) gin.HandlerFunc {
	store := cookie.NewStore([]byte(sessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	return sessions.Sessions("node-registrar-session", store)
}

// Helper functions
func abortWithError(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, gin.H{"error": msg})
}

func handleDatabaseError(c *gin.Context, err error) {
	if errors.Is(err, db.ErrRecordNotFound) {
		abortWithError(c, http.StatusNotFound, "Account not found")
	} else {
		abortWithError(c, http.StatusInternalServerError, "Database error")
	}
}
