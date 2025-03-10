package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

var ErrorNodeNotFound = fmt.Errorf("failed to get requested node from node registrar")

func (c *RegistrarClient) RegisterNode(
	farmID uint64,
	twinID uint64,
	interfaces []Interface,
	location Location,
	resources Resources,
	serialNumber string,
	secureBoot,
	virtualized bool,
) (nodeID uint64, err error) {
	return c.registerNode(farmID, twinID, interfaces, location, resources, serialNumber, secureBoot, virtualized)
}

func (c *RegistrarClient) UpdateNode(opts ...UpdateNodeOpts) (err error) {
	return c.updateNode(opts)
}

func (c *RegistrarClient) ReportUptime(report UptimeReport) (err error) {
	return c.reportUptime(report)
}

func (c *RegistrarClient) GetNode(id uint64) (node Node, err error) {
	return c.getNode(id)
}

func (c *RegistrarClient) GetNodeByTwinID(id uint64) (node Node, err error) {
	return c.getNodeByTwinID(id)
}

func (c *RegistrarClient) ListNodes(opts ...ListNodeOpts) (nodes []Node, err error) {
	return c.listNodes(opts)
}

type nodeCfg struct {
	nodeID        uint64
	farmID        uint64
	twinID        uint64
	status        string
	healthy       bool
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

func (c *RegistrarClient) registerNode(
	farmID uint64,
	twinID uint64,
	interfaces []Interface,
	location Location,
	resources Resources,
	serialNumber string,
	secureBoot,
	virtualized bool,
) (nodeID uint64, err error) {
	err = c.ensureTwinID()
	if err != nil {
		return nodeID, errors.Wrap(err, "failed to ensure twin id")
	}
	url, err := url.JoinPath(c.baseURL, "nodes")
	if err != nil {
		return nodeID, errors.Wrap(err, "failed to construct registrar url")
	}

	data := Node{
		FarmID:       farmID,
		TwinID:       twinID,
		Location:     location,
		Resources:    resources,
		Interfaces:   interfaces,
		SecureBoot:   secureBoot,
		Virtualized:  virtualized,
		SerialNumber: serialNumber,
	}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(data)
	if err != nil {
		return nodeID, errors.Wrap(err, "failed to encode request body")
	}

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

	c.nodeID = result.NodeID
	return result.NodeID, err
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

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(node)
	if err != nil {
		return errors.Wrap(err, "failed to encode request body")
	}

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

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return errors.Wrapf(err, "failed to update node with twin id %d with status code %s", c.twinID, resp.Status)
	}
	defer resp.Body.Close()

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

	var body bytes.Buffer

	err = json.NewEncoder(&body).Encode(report)
	if err != nil {
		return errors.Wrap(err, "failed to encode request body")
	}

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

	if resp.StatusCode != http.StatusCreated {
		err = parseResponseError(resp.Body)
		return errors.Wrapf(err, "failed to update node uptime for node with id %d with status code %s", c.nodeID, resp.Status)
	}
	defer resp.Body.Close()

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

	if resp.StatusCode == http.StatusNotFound {
		return node, ErrorNodeNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return node, errors.Wrapf(err, "failed to get node with status code %s", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&node)
	if err != nil {
		return
	}

	return
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

	if resp.StatusCode == http.StatusNotFound {
		return nodes, ErrorNodeNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return nodes, errors.Wrapf(err, "failed to list nodes with with status code %s", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&nodes)
	if err != nil {
		return
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
		nodeID:  0,
		twinID:  0,
		farmID:  0,
		status:  "",
		healthy: false,
		size:    50,
		page:    1,
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

	data["size"] = cfg.size
	data["page"] = cfg.page

	return data
}
