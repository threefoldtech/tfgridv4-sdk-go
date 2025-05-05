package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var ErrorNodeNotFound = fmt.Errorf("failed to get requested node from node registrar")

// RegisterNode register physical/virtual nodes with on TFGrid.
func (c *RegistrarClient) RegisterNode(node Node) (nodeID uint64, err error) {
	return c.registerNode(node)
}

// UpdateNode update node configuration (farmID, interfaces, resources, location, secureBoot, virtualized).
func (c *RegistrarClient) UpdateNode(opts ...UpdateNodeOpts) (err error) {
	return c.updateNode(opts)
}

// ReportUptime update node Uptime.
func (c *RegistrarClient) ReportUptime(report UptimeReport) (err error) {
	return c.reportUptime(report)
}

// GetNode gets registered node details using nodeID
func (c *RegistrarClient) GetNode(id uint64) (node Node, err error) {
	return c.getNode(id)
}

// GetNodeByTwinID gets registered node details using twinID
func (c *RegistrarClient) GetNodeByTwinID(id uint64) (node Node, err error) {
	return c.getNodeByTwinID(id)
}

// ListNodes lists registered nodes details using (nodeID, twinID, farmID).
func (c *RegistrarClient) ListNodes(opts ...ListNodeOpts) (nodes []Node, err error) {
	return c.listNodes(opts)
}

type nodeCfg struct {
	nodeID        uint64
	farmID        uint64
	twinID        uint64
	status        string
	healthy       bool
	online        *bool
	lastSeen      *int64
	Location      Location
	Resources     Resources
	Interfaces    []Interface
	SecureBoot    bool
	Virtualized   bool
	SerialNumber  string
	UptimeReports []UptimeReport
	Approved      bool
	page          uint32
	size          uint32
}

type (
	ListNodeOpts   func(*nodeCfg)
	UpdateNodeOpts func(*nodeCfg)
)

func ListNodesWithNodeID(id uint64) ListNodeOpts {
	return func(n *nodeCfg) {
		n.nodeID = id
	}
}

func ListNodesWithFarmID(id uint64) ListNodeOpts {
	return func(n *nodeCfg) {
		n.farmID = id
	}
}

func ListNodesWithStatus(status string) ListNodeOpts {
	return func(n *nodeCfg) {
		n.status = status
	}
}

func ListNodesWithHealthy() ListNodeOpts {
	return func(n *nodeCfg) {
		n.healthy = true
	}
}

func ListNodesWithOnline(online bool) ListNodeOpts {
	return func(n *nodeCfg) {
		n.online = &online
	}
}

func ListNodesWithLastSeen(minutes int64) ListNodeOpts {
	return func(n *nodeCfg) {
		n.lastSeen = &minutes
	}
}

func ListNodesWithTwinID(id uint64) ListNodeOpts {
	return func(n *nodeCfg) {
		n.twinID = id
	}
}

func ListNodesWithPage(page uint32) ListNodeOpts {
	return func(n *nodeCfg) {
		n.page = page
	}
}

func ListNodesWithSize(size uint32) ListNodeOpts {
	return func(n *nodeCfg) {
		n.size = size
	}
}

func UpdateNodesWithFarmID(id uint64) UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.farmID = id
	}
}

func UpdateNodesWithInterfaces(interfaces []Interface) UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.Interfaces = interfaces
	}
}

func UpdateNodesWithLocation(location Location) UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.Location = location
	}
}

func UpdateNodesWithResources(resources Resources) UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.Resources = resources
	}
}

func UpdateNodesWithSerialNumber(serialNumbe string) UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.SerialNumber = serialNumbe
	}
}

func UpdateNodesWithSecureBoot() UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.SecureBoot = true
	}
}

func UpdateNodesWithVirtualized() UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.Virtualized = true
	}
}

func UpdateNodeWithStatus(status string) UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.status = status
	}
}

func UpdateNodeWithHealthy() UpdateNodeOpts {
	return func(n *nodeCfg) {
		n.healthy = true
	}
}

func (c *RegistrarClient) registerNode(node Node) (nodeID uint64, err error) {
	err = c.ensureTwinID()
	if err != nil {
		return nodeID, errors.Wrap(err, "failed to ensure twin id")
	}

	url, err := url.JoinPath(c.baseURL, "nodes")
	if err != nil {
		return nodeID, errors.Wrap(err, "failed to construct registrar url")
	}

	handler := func(body bytes.Buffer) (uint64, error) {
		req, err := http.NewRequest("POST", url, &body)
		if err != nil {
			return nodeID, errors.Wrap(err, "failed to construct http request to the registrar")
		}

		authHeader, err := c.signRequest(time.Now().Unix())
		if err != nil {
			return nodeID, errors.Wrap(err, "failed to sign request")
		}
		req.Header.Set("X-Auth", authHeader)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nodeID, errors.Wrap(err, "failed to send request to registrer the node")
		}

		if resp == nil {
			return 0, errors.New("no response received")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			err = parseResponseError(resp.Body)
			return 0, errors.Wrapf(err, "failed to create node on the registrar with status code %s", resp.Status)
		}

		result := struct {
			NodeID uint64 `json:"node_id"`
		}{}

		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return 0, errors.Wrap(err, "failed to decode response body")
		}

		return result.NodeID, nil
	}

	var body bytes.Buffer
	if err = json.NewEncoder(&body).Encode(node); err != nil {
		return 0, err
	}

	nodeID, err = handler(body)
	if err != nil {
		// try node with old interface format
		body, err := createRequestBodyWithOldInterfaceFormat(node)
		if err != nil {
			return nodeID, err
		}

		nodeID, err = handler(body)
		if err != nil {
			return nodeID, err
		}
	}

	c.nodeID = nodeID
	return nodeID, nil
}

func (c *RegistrarClient) updateNode(opts []UpdateNodeOpts) (err error) {
	err = c.ensureNodeID()
	if err != nil {
		return err
	}

	node, err := c.getNode(c.nodeID)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.baseURL, "nodes", fmt.Sprint(c.nodeID))
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	node = c.parseUpdateNodeOpts(node, opts)

	handler := func(body bytes.Buffer) error {
		req, err := http.NewRequest("PATCH", url, &body)
		if err != nil {
			return errors.Wrap(err, "failed to construct http request to the registrar")
		}

		authHeader, err := c.signRequest(time.Now().Unix())
		if err != nil {
			return errors.Wrap(err, "failed to sign request")
		}
		req.Header.Set("X-Auth", authHeader)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "failed to send request to update node")
		}

		if resp == nil {
			return errors.New("no response received")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = parseResponseError(resp.Body)
			return errors.Wrapf(err, "failed to update node with twin id %d with status code %s", c.twinID, resp.Status)
		}
		return nil
	}

	var body bytes.Buffer
	if err = json.NewEncoder(&body).Encode(node); err != nil {
		return err
	}

	if err := handler(body); err != nil {
		// try update with old interface format
		node, err := createRequestBodyWithOldInterfaceFormat(node)
		if err != nil {
			return err
		}

		if err := handler(node); err != nil {
			return err
		}
	}

	return
}

func (c *RegistrarClient) reportUptime(report UptimeReport) (err error) {
	err = c.ensureNodeID()
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.baseURL, "nodes", fmt.Sprint(c.nodeID), "uptime")
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	handler := func(body bytes.Buffer) error {
		req, err := http.NewRequest("POST", url, &body)
		if err != nil {
			return errors.Wrap(err, "failed to construct http request to the registrar")
		}

		authHeader, err := c.signRequest(time.Now().Unix())
		if err != nil {
			return errors.Wrap(err, "failed to sign request")
		}
		req.Header.Set("X-Auth", authHeader)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "failed to send request to update uptime of the node")
		}

		if resp == nil {
			return errors.New("no response received")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			err = parseResponseError(resp.Body)
			return errors.Wrapf(err, "failed to update node uptime for node with id %d with status code %s", c.nodeID, resp.Status)
		}
		return nil
	}

	var body bytes.Buffer
	if err = json.NewEncoder(&body).Encode(report); err != nil {
		return errors.Wrap(err, "failed to encode request body")
	}

	if err = handler(body); err != nil {
		// try old report format time.Duration and time.Time
		old := struct {
			Uptime    time.Duration
			Timestamp time.Time
		}{
			Uptime:    time.Duration(report.Uptime),
			Timestamp: time.Now(),
		}
		var body bytes.Buffer
		if err = json.NewEncoder(&body).Encode(old); err != nil {
			return errors.Wrap(err, "failed to encode request body")
		}

		if err = handler(body); err != nil {
			return err
		}
	}

	return
}

func (c *RegistrarClient) getNode(id uint64) (node Node, err error) {
	url, err := url.JoinPath(c.baseURL, "nodes", fmt.Sprint(id))
	if err != nil {
		return node, errors.Wrap(err, "failed to construct registrar url")
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return
	}

	if resp == nil {
		return node, errors.New("no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return node, ErrorNodeNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return node, errors.Wrapf(err, "failed to get node with status code %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(bodyBytes, &node)
	if err == nil {
		return
	}

	return parseResponseBodyToNewInterfaceFormat(bodyBytes)
}

func (c *RegistrarClient) getNodeByTwinID(id uint64) (node Node, err error) {
	nodes, err := c.ListNodes(ListNodesWithTwinID(id))
	if err != nil {
		return
	}

	if len(nodes) == 0 {
		return node, ErrorNodeNotFound
	}

	return nodes[0], nil
}

func (c *RegistrarClient) listNodes(opts []ListNodeOpts) (nodes []Node, err error) {
	url, err := url.JoinPath(c.baseURL, "nodes")
	if err != nil {
		return nodes, errors.Wrap(err, "failed to construct registrar url")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nodes, errors.Wrap(err, "failed to construct http request to the registrar")
	}

	q := req.URL.Query()
	data := parseListNodeOpts(opts)

	for key, val := range data {
		q.Add(key, fmt.Sprint(val))
	}

	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}

	if resp == nil {
		return nodes, errors.New("no response received")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nodes, ErrorNodeNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return nodes, errors.Wrapf(err, "failed to list nodes with with status code %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(bodyBytes, &nodes)
	if err == nil {
		return
	}

	var rawNodes []interface{}
	err = json.Unmarshal(bodyBytes, &rawNodes)
	if err != nil {
		return
	}

	for _, node := range rawNodes {
		nodeBytes, err := json.Marshal(node)
		if err != nil {
			return nodes, err
		}

		node, err := parseResponseBodyToNewInterfaceFormat(nodeBytes)
		if err != nil {
			return nodes, err
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (c *RegistrarClient) ensureNodeID() error {
	if c.nodeID != 0 {
		return nil
	}

	err := c.ensureTwinID()
	if err != nil {
		return err
	}

	node, err := c.getNodeByTwinID(c.twinID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the node id, registrar client was set up with a normal account not a node")
	}

	c.nodeID = node.NodeID
	return nil
}

func (c *RegistrarClient) parseUpdateNodeOpts(node Node, opts []UpdateNodeOpts) Node {
	cfg := nodeCfg{
		farmID:      0,
		Location:    Location{},
		Resources:   Resources{},
		Interfaces:  []Interface{},
		SecureBoot:  false,
		Virtualized: false,
		Approved:    false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.farmID != 0 {
		node.FarmID = cfg.farmID
	}

	if !reflect.DeepEqual(cfg.Location, Location{}) {
		node.Location = cfg.Location
	}

	if !reflect.DeepEqual(cfg.Resources, Resources{}) {
		node.Resources = cfg.Resources
	}

	if len(cfg.Interfaces) != 0 {
		node.Interfaces = cfg.Interfaces
	}

	return node
}

func parseListNodeOpts(opts []ListNodeOpts) map[string]any {
	cfg := nodeCfg{
		nodeID:   0,
		twinID:   0,
		farmID:   0,
		status:   "",
		healthy:  false,
		online:   nil,
		lastSeen: nil,
		size:     50,
		page:     1,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	data := map[string]any{}

	if cfg.nodeID != 0 {
		data["node_id"] = cfg.nodeID
	}

	if cfg.twinID != 0 {
		data["twin_id"] = cfg.twinID
	}

	if cfg.farmID != 0 {
		data["farm_id"] = cfg.farmID
	}

	if len(cfg.status) != 0 {
		data["status"] = cfg.status
	}

	if cfg.healthy {
		data["healthy"] = cfg.healthy
	}

	if cfg.online != nil {
		data["online"] = *cfg.online
	}

	if cfg.lastSeen != nil {
		data["last_seen"] = *cfg.lastSeen
	}

	data["size"] = cfg.size
	data["page"] = cfg.page

	return data
}

//
// func migrateInterfaces(nodeMap map[string]any) (Node, error) {
// 	type oldInterface struct {
// 		Name string `json:"name"`
// 		Mac  string `json:"mac"`
// 		IPs  string `json:"ips"`
// 	}
//
// 	rawInterfaces, ok := nodeMap["interfaces"].([]interface{})
// 	if !ok {
// 		return Node{}, fmt.Errorf("interfaces is not a list")
// 	}
//
// 	interfaceBytes, err := json.Marshal(rawInterfaces)
// 	if err != nil {
// 		return Node{}, fmt.Errorf("failed to marshal interfaces: %w", err)
// 	}
//
// 	var old []oldInterface
// 	if err := json.Unmarshal(interfaceBytes, &old); err != nil {
// 		return Node{}, errors.New("node interfaces doesn't implement either old or new Interface struct")
// 	}
// 	var newInterfaces []Interface
// 	for _, ifs := range old {
// 		ips := strings.Split(ifs.IPs, "/")
// 		newI := Interface{
// 			Name: ifs.Name,
// 			Mac:  ifs.Mac,
// 			IPs:  ips,
// 		}
// 		newInterfaces = append(newInterfaces, newI)
// 	}
// 	nodeMap["interfaces"] = newInterfaces
//
// 	encodedNode, err := json.Marshal(nodeMap)
// 	if err != nil {
// 		return Node{}, err
// 	}
//
// 	var updatedNode Node
// 	err = json.Unmarshal(encodedNode, updatedNode)
//
// 	return updatedNode, err
// }

/*
*
*
*
*
*
*
*
*
 */

type oldInterfaceFormat struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
	IPs  string `json:"ips"`
}

func createRequestBodyWithOldInterfaceFormat(node Node) (body bytes.Buffer, err error) {
	var oldInterfaces []oldInterfaceFormat
	interfaces := node.Interfaces
	for _, i := range interfaces {
		old := oldInterfaceFormat{
			Name: i.Name,
			Mac:  i.Mac,
			IPs:  strings.Join(i.IPs, "/"),
		}
		oldInterfaces = append(oldInterfaces, old)
	}

	nodeBytes, err := json.Marshal(node)
	if err != nil {
		return
	}

	var nodeMap map[string]any
	err = json.Unmarshal(nodeBytes, &nodeMap)
	if err != nil {
		return
	}
	nodeMap["interfaces"] = oldInterfaces

	err = json.NewEncoder(&body).Encode(node)
	if err != nil {
		return
	}

	return
}

func parseResponseBodyToNewInterfaceFormat(nodeBytes []byte) (Node, error) {
	oldFormatNode := struct {
		Interfaces []oldInterfaceFormat `json:"interfaces"`
	}{}

	err := json.Unmarshal(nodeBytes, &oldFormatNode)
	if err != nil {
		return Node{}, err
	}

	var newFormat []Interface
	for _, i := range oldFormatNode.Interfaces {
		ips := strings.Split(i.IPs, "/")
		newFormat = append(newFormat, Interface{
			Name: i.Name,
			Mac:  i.Mac,
			IPs:  ips,
		})
	}

	var nodeMap map[string]any
	err = json.Unmarshal(nodeBytes, &nodeMap)
	if err != nil {
		return Node{}, err
	}
	nodeMap["interfaces"] = newFormat

	encodedNode, err := json.Marshal(nodeMap)
	if err != nil {
		return Node{}, err
	}

	var node Node
	err = json.Unmarshal(encodedNode, &node)
	if err != nil {
		return Node{}, err
	}

	return node, nil
}

// nodeBytes, err := json.Marshal(nodeMap)
// if err != nil {
// 	return Node{}, err
// }
//
