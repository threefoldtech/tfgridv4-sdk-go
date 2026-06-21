package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/pkg/db"
)

const (
	MaxTimestampDelta                    = 2 * time.Second
	UptimeReportTimestampHintDrift int64 = 60
	OnlineCutoffTime                     = 40 * time.Minute

	// Database field size limits
	MaxFarmNameLength       = 40
	MaxStellarAddressLength = 56
	MaxKeySize              = 50

	// Default pagination
	DefaultPageSize = 10

	// Time constants
	DefaultOnlineCutoffMinutes = 40
)

// @title Node Registrar API
// @version 1.0
// @description API for managing TFGrid node registration
// @BasePath /api/v1

// @Summary List farms
// @Description Get a list of farms with optional filters
// @Tags farms
// @Accept json
// @Produce json
// @Param farm_name query string false "Filter by farm name"
// @Param farm_id query int false "Filter by farm ID"
// @Param twin_id query int false "Filter by twin ID"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Results per page" default(10)
// @Success 200 {object} []db.Farm "List of farms"
// @Failure 400 {object} map[string]any "Bad request"
// @Router /farms [get]
func (s Server) listFarmsHandler(c *gin.Context) {
	var filter db.FarmFilter
	limit := db.DefaultLimit()

	err := parseQueryParams(c, &limit, &filter)
	if err != nil {
		Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	farms, err := s.db.ListFarms(filter, limit)
	if err != nil {
		Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	Response(c, http.StatusOK, "Farms are listed successfully", farms)

}

// @Summary Get farm details
// @Description Get details for a specific farm
// @Tags farms
// @Accept json
// @Produce json
// @Param farm_id path int true "Farm ID"
// @Success 200 {object} db.Farm "Farm details"
// @Failure 400 {object} map[string]any "Invalid farm ID"
// @Failure 404 {object} map[string]any "Farm not found"
// @Router /farms/{farm_id} [get]
func (s Server) getFarmHandler(c *gin.Context) {
	farmID := c.Param("farm_id")

	id, err := strconv.ParseUint(farmID, 10, 64)
	if err != nil {
		Response(c, http.StatusBadRequest, fmt.Sprintf("Invalid farm_id: %v", err.Error()), nil)
		return
	}

	farm, err := s.db.GetFarm(id)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			Response(c, http.StatusNotFound, err.Error(), nil)
			return
		}

		Response(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	Response(c, http.StatusOK, "Farm retrieved successfully", farm)
}

// @Summary Create new farm
// @Description Create a new farm entry
// @Tags farms
// @Accept json
// @Produce json
// @Param X-Auth header string true "Authentication format: Base64(<unix_timestamp>:<twin_id>):Base64(signature)"
// @Param farm body db.Farm true "Farm creation data"
// @Success 201 {object} map[string]uint64 "'farm_id': farmID"]
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 409 {object} map[string]any "Farm already exists"
// @Router /farms [post]
func (s Server) createFarmHandler(c *gin.Context) {
	var farm db.Farm

	if err := c.ShouldBindJSON(&farm); err != nil {
		Response(c, http.StatusBadRequest, fmt.Sprintf("failed to parse farm info: %v", err.Error()), nil)
		return
	}

	ensureOwner(c, farm.TwinID)
	if c.IsAborted() {
		return
	}

	farmID, err := s.db.CreateFarm(farm)
	if err != nil {
		if errors.Is(err, db.ErrRecordAlreadyExists) {
			Response(c, http.StatusConflict, err.Error(), nil)
			return
		}

		Response(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	Response(c, http.StatusCreated, "Farm created successfully", gin.H{"farm_id": farmID})
}

type UpdateFarmRequest struct {
	FarmName       string `json:"farm_name" binding:"max=40"`
	StellarAddress string `json:"stellar_address" binding:"startswith=G,len=56,alphanum,uppercase"`
}

// @Summary Update farm
// @Description Update existing farm details
// @Tags farms
// @Accept json
// @Produce json
// @Param X-Auth header string true "Authentication format: Base64(<unix_timestamp>:<twin_id>):Base64(signature)"
// @Param farm_id path int true "Farm ID"
// @Param request body UpdateFarmRequest true "Farm update data"
// @Success 200 {object} map[string]any "Farm updated successfully"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 404 {object} map[string]any "Farm not found"
// @Router /farms/{farm_id} [patch]
func (s Server) updateFarmHandler(c *gin.Context) {
	var req UpdateFarmRequest
	farmID := c.Param("farm_id")

	id, err := strconv.ParseUint(farmID, 10, 64)
	if err != nil {
		Response(c, http.StatusBadRequest, fmt.Sprintf("Invalid farm_id: %v", err.Error()), nil)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Response(c, http.StatusBadRequest, fmt.Sprintf("failed to parse farm info: %v", err.Error()), nil)
		return
	}

	existingFarm, err := s.db.GetFarm(id)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			Response(c, http.StatusNotFound, "Farm not found", nil)
			return
		}
		Response(c, http.StatusInternalServerError, "Database error", nil)
		return
	}

	ensureOwner(c, existingFarm.TwinID)
	if c.IsAborted() {
		return
	}

	req.FarmName = strings.TrimSpace(req.FarmName)
	req.StellarAddress = strings.TrimSpace(req.StellarAddress)

	// No need to hit DB if new farm name is same as the old one
	if (len(req.FarmName) != 0 && existingFarm.FarmName != req.FarmName) ||
		(len(req.StellarAddress) != 0 && existingFarm.StellarAddress != req.StellarAddress) {
		err = s.db.UpdateFarm(id, req.FarmName, req.StellarAddress)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				Response(c, http.StatusNotFound, err.Error(), nil)
				return
			}

			Response(c, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	Response(c, http.StatusOK, "Farm was updated successfully", nil)
}

// @Summary List nodes
// @Description Get a list of nodes with optional filters
// @Tags nodes
// @Accept json
// @Produce json
// @Param node_id query int false "Filter by node ID"
// @Param farm_id query int false "Filter by farm ID"
// @Param twin_id query int false "Filter by twin ID"
// @Param status query string false "Filter by status"
// @Param healthy query bool false "Filter by health status"
// @Param online query bool false "Filter by online status (true = online, false = offline)"
// @Param last_seen query int false "Filter nodes last seen within this many minutes"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Results per page" default(10)
// @Success 200 {object} []db.Node "List of nodes with online status"
// @Failure 400 {object} map[string]any "Bad request"
// @Router /nodes [get]
func (s Server) listNodesHandler(c *gin.Context) {
	var filter db.NodeFilter
	limit := db.DefaultLimit()

	err := parseQueryParams(c, &limit, &filter)
	if err != nil {
		Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	nodes, err := s.db.ListNodes(filter, limit)
	if err != nil {
		Response(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Set online status for each node
	cutoffTime := time.Now().Add(-OnlineCutoffTime)
	for i := range nodes {
		nodes[i].Online = !nodes[i].LastSeen.IsZero() && nodes[i].LastSeen.After(cutoffTime)
	}

	Response(c, http.StatusOK, "Nodes are listed successfully", nodes)
}

// @Summary Get node details
// @Description Get details for a specific node
// @Tags nodes
// @Accept json
// @Produce json
// @Param node_id path int true "Node ID"
// @Success 200 {object} db.Node "Node details with online status and last_seen information"
// @Failure 400 {object} map[string]any "Invalid node ID"
// @Failure 404 {object} map[string]any "Node not found"
// @Router /nodes/{node_id} [get]
func (s Server) getNodeHandler(c *gin.Context) {
	nodeID := c.Param("node_id")

	id, err := strconv.ParseUint(nodeID, 10, 64)
	if err != nil {
		Response(c, http.StatusBadRequest, "Invalid node id", nil)
		return
	}

	node, err := s.db.GetNode(id)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			Response(c, http.StatusNotFound, err.Error(), nil)
			return
		}

		Response(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	// Determine if the node is online (has sent an uptime report in the last 30 minutes)
	cutoffTime := time.Now().Add(-OnlineCutoffTime)
	node.Online = !node.LastSeen.IsZero() && node.LastSeen.After(cutoffTime)

	Response(c, http.StatusOK, "Node retrieved successfully", node)
}

type NodeRegistrationRequest struct {
	TwinID       uint64         `json:"twin_id" binding:"required,min=1"`
	FarmID       uint64         `json:"farm_id" binding:"required,min=1"`
	Resources    db.Resources   `json:"resources" binding:"required"`
	Location     db.Location    `json:"location" binding:"required"`
	Interfaces   []db.Interface `json:"interfaces" binding:"required"`
	SecureBoot   bool           `json:"secure_boot"`
	Virtualized  bool           `json:"virtualized"`
	SerialNumber string         `json:"serial_number" binding:"required"`
}

// @Summary Register new node
// @Description Register a new node in the system
// @Tags nodes
// @Accept json
// @Produce json
// @Param X-Auth header string true "Authentication format: Base64(<unix_timestamp>:<twin_id>):Base64(signature)"
// @Param request body NodeRegistrationRequest true "Node registration data"
// @Success 201 {object} map[string]uint64 "'node_id': nodeID"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 409 {object} map[string]any "Node already exists"
// @Router /nodes [post]
func (s Server) registerNodeHandler(c *gin.Context) {
	var req NodeRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	ensureOwner(c, req.TwinID)
	if c.IsAborted() {
		return
	}

	node := db.Node{
		TwinID:       req.TwinID,
		FarmID:       req.FarmID,
		Resources:    req.Resources,
		Location:     req.Location,
		Interfaces:   req.Interfaces,
		SecureBoot:   req.SecureBoot,
		Virtualized:  req.Virtualized,
		SerialNumber: req.SerialNumber,
		Approved:     false, // Default to unapproved awaiting farmer approval
	}

	nodeID, err := s.db.RegisterNode(node)
	if err != nil {
		if errors.Is(err, db.ErrRecordAlreadyExists) {
			Response(c, http.StatusConflict, err.Error(), nil)
			return
		}

		Response(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	Response(c, http.StatusCreated, "Node registered successfully", gin.H{"node_id": nodeID})
}

type UpdateNodeRequest struct {
	FarmID       uint64         `json:"farm_id" binding:"required,min=1"`
	Resources    db.Resources   `json:"resources" binding:"required"`
	Location     db.Location    `json:"location" binding:"required"`
	Interfaces   []db.Interface `json:"interfaces" binding:"required"`
	SecureBoot   bool           `json:"secure_boot"`
	Virtualized  bool           `json:"virtualized"`
	SerialNumber string         `json:"serial_number" binding:"required"`
}

// @Summary Update node
// @Description Update existing node details
// @Tags nodes
// @Accept json
// @Produce json
// @Param X-Auth header string true "Authentication format: Base64(<unix_timestamp>:<twin_id>):Base64(signature)"
// @Param node_id path int true "Node ID"
// @Param request body UpdateNodeRequest true "Node update data"
// @Success 200 {object} map[string]any "Node updated successfully"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 404 {object} map[string]any "Node not found"
// @Router /nodes/{node_id} [patch]
func (s *Server) updateNodeHandler(c *gin.Context) {
	nodeID, err := strconv.ParseUint(c.Param("node_id"), 10, 64)
	if err != nil {
		AbortResponse(c, http.StatusBadRequest, "invalid node ID", nil)
		return
	}

	existingNode, err := s.db.GetNode(nodeID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			AbortResponse(c, http.StatusNotFound, "node not found", nil)
			return
		}
		AbortResponse(c, http.StatusInternalServerError, "database error", nil)
		return
	}

	ensureOwner(c, existingNode.TwinID)
	if c.IsAborted() {
		return
	}

	var req UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		AbortResponse(c, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	updatedNode := db.Node{
		FarmID:       req.FarmID,
		Resources:    req.Resources,
		Location:     req.Location,
		Interfaces:   req.Interfaces,
		SecureBoot:   req.SecureBoot,
		Virtualized:  req.Virtualized,
		SerialNumber: req.SerialNumber,
	}

	if req.FarmID != existingNode.FarmID {
		updatedNode.Approved = false
	}

	if err := s.db.UpdateNode(nodeID, updatedNode); err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			AbortResponse(c, http.StatusNotFound, "node not found", nil)
			return
		}
		AbortResponse(c, http.StatusInternalServerError, "failed to update node", nil)
		return
	}

	Response(c, http.StatusOK, "node updated successfully", nil)
}

type UptimeReportRequest struct {
	Uptime    uint64 `json:"uptime" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
}

// @Summary Report node uptime
// @Description Submit uptime report for a node
// @Tags nodes
// @Accept json
// @Produce json
// @Param X-Auth header string true "Authentication format: Base64(<unix_timestamp>:<twin_id>):Base64(signature)"
// @Param node_id path int true "Node ID"
// @Param request body UptimeReportRequest true "Uptime report data"
// @Success 201 {object} map[string]any "Uptime reported successfully"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 404 {object} map[string]any "Node not found"
// @Router /nodes/{node_id}/uptime [post]
func (s *Server) uptimeReportHandler(c *gin.Context) {
	nodeID := c.Param("node_id")

	id, err := strconv.ParseUint(nodeID, 10, 64)
	if err != nil {
		Response(c, http.StatusBadRequest, "Invalid node id", nil)
		return
	}

	var req UptimeReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Get node
	node, err := s.db.GetNode(id)
	if err != nil {
		Response(c, http.StatusNotFound, "node not found", nil)
		return
	}

	ensureOwner(c, node.TwinID)
	if c.IsAborted() {
		return
	}
	// Detect restarts
	// Validate report timing (40min ± 5min window)
	// Maybe aggregate reports here and store total uptime?
	// The total uptime should accumulate unless the node restarts, which is detected when the reported uptime is less than the previous value.

	// Ensuring the timestamp_hint is within an Acceptable Range
	err = validateTimestampHint(req.Timestamp)
	if err != nil {
		// include the error message
		Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Create report record
	report := &db.UptimeReport{
		NodeID:    id,
		Duration:  time.Duration(req.Uptime) * time.Second,
		Timestamp: time.Unix(req.Timestamp, 0).UTC(),
	}

	// Create report record and Update node LastSeen(the timestamp of the last report)
	// It's up to the clients to determine if the node is online based on the reporting interval and allowable window.
	err = s.db.CreateUptimeReport(report)
	if err != nil {
		Response(c, http.StatusInternalServerError, "failed to process uptime report", nil)
		return
	}

	Response(c, http.StatusCreated, "uptime reported successfully", nil)
}

func parseQueryParams(c *gin.Context, types_ ...interface{}) error {
	for _, type_ := range types_ {
		if err := c.ShouldBindQuery(type_); err != nil {
			return fmt.Errorf("failed to bind query params to %T: %w", type_, err)
		}
	}
	return nil
}

// AccountRequest represents the request body for account operations
type AccountCreationRequest struct {
	Timestamp int64  `json:"timestamp" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"` // base64 encoded
	// the registrar expect a signature of a message with format `timestampStr:publicKeyBase64`
	// - signature format: base64(ed25519_or_sr22519_signature)
	Signature string   `json:"signature" binding:"required"`
	Relays    []string `json:"relays,omitempty"`
	RMBEncKey string   `json:"rmb_enc_key,omitempty"`
}

// @Summary Create new account
// @Description Create a new twin account with cryptographic verification
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body AccountCreationRequest true "Account creation data"
// @Success 201 {object} db.Account "Created account details"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 409 {object} map[string]any "Account already exists"
// @Router /accounts [post]
func (s *Server) createAccountHandler(c *gin.Context) {
	var req AccountCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Response(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Validate public key format
	if !isValidPublicKey(req.PublicKey) {
		Response(c, http.StatusBadRequest, "invalid public key format", nil)
		return
	}

	// Verify timestamp is within acceptable window
	now := time.Now()
	requestTime := time.Unix(req.Timestamp, 0)
	delta := now.Sub(requestTime)

	if delta < -MaxTimestampDelta || delta > MaxTimestampDelta {
		Response(c, http.StatusBadRequest, "timestamp outside acceptable window", gin.H{
			"server_time": now.Unix(),
		})
		return
	}

	// Create challenge using timestamp and public key
	// Challenge is uniquely tied to both the timestamp and public key
	// Prevents replay attacks, still no state management required
	challenge := []byte(fmt.Sprintf("%d:%s", req.Timestamp, req.PublicKey))

	// Decode public key from base64
	publicKeyBytes, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil {
		Response(c, http.StatusBadRequest, "invalid public key format", nil)
		return
	}
	// Decode signature from base64
	signatureBytes, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		Response(c, http.StatusBadRequest, fmt.Sprintf("invalid signature format: %v", err), nil)
		return
	}
	// Verify signature of the challenge
	err = verifySignature(publicKeyBytes, challenge, signatureBytes)
	if err != nil {
		Response(c, http.StatusUnauthorized, fmt.Sprintf("signature verification error: %v", err), nil)
		return
	}

	// Now we can create new account
	account := &db.Account{
		PublicKey: req.PublicKey,
		Relays:    req.Relays,
		RMBEncKey: req.RMBEncKey,
	}

	if err := s.db.CreateAccount(account); err != nil {
		if errors.Is(err, db.ErrRecordAlreadyExists) {
			Response(c, http.StatusConflict, "account with this public key already exists", nil)
			return
		}
		Response(c, http.StatusInternalServerError, "failed to create account", nil)
		return
	}

	Response(c, http.StatusCreated, "Account created successfully", account)
}

type UpdateAccountRequest struct {
	Relays    pq.StringArray `json:"relays" swaggertype:"array,string"`
	RMBEncKey string         `json:"rmb_enc_key"`
}

// updateAccountHandler updates an account's relays and RMB encryption key
// @Summary Update account details
// @Description Updates an account's relays and RMB encryption key
// @Tags accounts
// @Accept json
// @Produce json
// @Param X-Auth header string true "Authentication format: Base64(<unix_timestamp>:<twin_id>):Base64(signature)"
// @Param twin_id path uint64 true "Twin ID of the account"
// @Param account body UpdateAccountRequest true "Account details to update"
// @Success 200 {object} map[string]any "Account updated successfully"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 404 {object} map[string]any "Account not found"
// @Router /accounts/{twin_id} [patch]
func (s *Server) updateAccountHandler(c *gin.Context) {
	twinID, err := strconv.ParseUint(c.Param("twin_id"), 10, 64)
	if err != nil {
		AbortResponse(c, http.StatusBadRequest, "invalid twin ID", nil)
		return
	}

	ensureOwner(c, twinID)
	if c.IsAborted() {
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		AbortResponse(c, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	err = s.db.UpdateAccount(twinID, req.Relays, req.RMBEncKey)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			AbortResponse(c, http.StatusNotFound, "account not found", nil)
			return
		}
		AbortResponse(c, http.StatusInternalServerError, "failed to update account", nil)
		return
	}

	Response(c, http.StatusOK, "account updated successfully", nil)
}

// getAccountHandler retrieves an account by twin ID or public key
// @Summary Retrieve an account by twin ID or public key
// @Description This endpoint retrieves an account by its twin ID or public key.
// @Tags accounts
// @Accept json
// @Produce json
// @Param twin_id query uint64 false "Twin ID of the account"
// @Param public_key query string false "Base64 decoded Public key of the account"
// @Success 200 {object} db.Account "Account details"
// @Failure 400 {object} map[string]any "Invalid request"
// @Failure 404 {object} map[string]any "Account not found"
// @Router /accounts [get]
func (s *Server) getAccountHandler(c *gin.Context) {
	twinIDParam := c.Query("twin_id")
	publicKeyParam := c.Query("public_key")

	// Validate only one parameter is provided
	if twinIDParam != "" && publicKeyParam != "" {
		AbortResponse(c, http.StatusBadRequest, "provide either twin_id or public_key, not both", nil)
		return
	}

	if twinIDParam == "" && publicKeyParam == "" {
		AbortResponse(c, http.StatusBadRequest, "must provide either twin_id or public_key parameter", nil)
		return
	}

	if twinIDParam != "" {
		twinID, err := strconv.ParseUint(twinIDParam, 10, 64)
		if err != nil {
			Response(c, http.StatusBadRequest, "invalid twin ID", nil)
			return
		}

		account, err := s.db.GetAccount(twinID)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				Response(c, http.StatusNotFound, "account not found", nil)
				return
			}
			Response(c, http.StatusInternalServerError, "failed to get account", nil)
			return
		}

		Response(c, http.StatusOK, "Account retrieved successfully", account)
		return
	}

	if publicKeyParam != "" {
		account, err := s.db.GetAccountByPublicKey(publicKeyParam)
		if err != nil {
			if errors.Is(err, db.ErrRecordNotFound) {
				Response(c, http.StatusNotFound, "account not found", nil)
				return
			}
			Response(c, http.StatusInternalServerError, "failed to get account", nil)
			return
		}
		Response(c, http.StatusOK, "Account retrieved successfully", account)
		return
	}
}

type ZOSVersionRequest struct {
	Version string `json:"version" binding:"required,base64"`
}

// @Summary Set ZOS Version
// @Description Sets the ZOS version
// @Tags ZOS
// @Accept json
// @Produce json
// @Param X-Auth header string true "Authentication format: Base64(<unix_timestamp>:<twin_id>):Base64(signature)"
// @Param body body ZOSVersionRequest true "Update ZOS Version Request"
// @Success 200 {object} map[string]any "OK"
// @Failure 400 {object} map[string]any "Bad Request"
// @Failure 401 {object} map[string]any "Unauthorized"
// @Failure 409 {object} map[string]any "Conflict"
// @Failure 500 {object} map[string]any "Internal Server Error"
// @Router /zos/version [put]
func (s *Server) setZOSVersionHandler(c *gin.Context) {
	ensureOwner(c, s.adminTwinID)
	if c.IsAborted() {
		return
	}

	var req ZOSVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		AbortResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := s.db.SetZOSVersion(req.Version); err != nil {
		if errors.Is(err, db.ErrVersionAlreadySet) {
			AbortResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		AbortResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	Response(c, http.StatusOK, "ZOS version set successfully", nil)
}

// @Summary Get ZOS Version
// @Description Gets the ZOS version
// @Tags ZOS
// @Produce json
// @Success 200 {object} string "zos version"
// @Failure 404 {object} map[string]any "Not Found"
// @Failure 500 {object} map[string]any "Internal Server Error"
// @Router /zos/version [get]
func (s *Server) getZOSVersionHandler(c *gin.Context) {
	version, err := s.db.GetZOSVersion()
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			AbortResponse(c, http.StatusNotFound, "zos version not set", nil)
			return
		}
		AbortResponse(c, http.StatusInternalServerError, "database error", nil)
		return
	}

	Response(c, http.StatusOK, "ZOS version retrieved successfully", version)
}

// Helper function to validate public key length
func isValidPublicKey(publicKeyBase64 string) bool {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return false
	}
	return len(publicKeyBytes) == PubKeySize
}

// Helper function to ensure the request is from the owner
func ensureOwner(c *gin.Context, twinID uint64) {
	// Retrieve twinID set by the authMiddleware
	authTwinID := c.Request.Context().Value(twinIDKey{})
	if authTwinID == nil {
		AbortResponse(c, http.StatusUnauthorized, "not authorized", nil)
		return
	}

	// Safe type assertion
	authID, ok := authTwinID.(uint64)
	if !ok {
		AbortResponse(c, http.StatusUnauthorized, "invalid authentication type", nil)
		return
	}

	// Ensure that the retrieved twinID equals to the passed twinID
	if authID != twinID || twinID == 0 {
		AbortResponse(c, http.StatusUnauthorized, "not authorized", nil)
		return
	}
}

// Helper function to validate timestamp hint
func validateTimestampHint(timestampHint int64) error {
	hintTime := time.Unix(timestampHint, 0)

	now := time.Now()

	// Calculate acceptable range
	maxDrift := time.Duration(UptimeReportTimestampHintDrift) * time.Second
	earliestAllowed := now.Add(-maxDrift)
	latestAllowed := now.Add(maxDrift)

	// Check if the hint is within the acceptable range
	if hintTime.Before(earliestAllowed) || hintTime.After(latestAllowed) {
		return fmt.Errorf("invalid timestamp hint: must be within ±%d seconds of the current time (%s)", UptimeReportTimestampHintDrift, now)
	}

	return nil
}
