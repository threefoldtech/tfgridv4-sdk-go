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

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
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

func (s *Server) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		stop := s.metrics.RecordDuration(s.metrics.HTTPRequestProcessingDuration, []string{c.Request.Method, c.Request.URL.Path})
		c.Next()

		stop()
		if c.Writer.Status() >= 500 {
			s.metrics.InternalErrors.Inc()
		}
		s.metrics.RecordCountVec(s.metrics.HTTPRequestsReceived, []string{c.Request.Method, c.Request.URL.Path, strconv.Itoa(c.Writer.Status())})
	}
}
