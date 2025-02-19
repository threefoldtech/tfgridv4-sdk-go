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

var ErrorNodeNotFround = fmt.Errorf("failed to get requested node from node regiatrar")

func (c RegistrarClient) RegisterNode(
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

func (c RegistrarClient) UpdateNode(opts ...UpdateNodeOpts) (err error) {
	return c.updateNode(opts)
}

func (c RegistrarClient) ReportUptime(report UptimeReport) (err error) {
	return c.reportUptime(report)
}

func (c RegistrarClient) GetNode(id uint64) (node Node, err error) {
	return c.getNode(id)
}

func (c RegistrarClient) GetNodeByTwinID(id uint64) (node Node, err error) {
	return c.getNodeByTwinID(id)
}

func (c RegistrarClient) ListNodes(opts ...ListNodeOpts) (nodes []Node, err error) {
	return c.ListNodes(opts)
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

func (c RegistrarClient) registerNode(
	farmID uint64,
	twinID uint64,
	interfaces []Interface,
	location Location,
	resources Resources,
	serialNumber string,
	secureBoot,
	virtualized bool,
) (nodeID uint64, err error) {
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

	req.Header.Set("X-Auth", c.signRequest(time.Now().Unix()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nodeID, errors.Wrap(err, "failed to send request to registrer the node")
	}

	if resp == nil || resp.StatusCode != http.StatusCreated {
		err = parseResponseError(resp.Body)
		return 0, errors.Wrapf(err, "failed to update node on the registrar with status code %s", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&nodeID)

	c.nodeID = nodeID
	return
}

func (c RegistrarClient) updateNode(opts []UpdateNodeOpts) (err error) {
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

	req.Header.Set("X-Auth", c.signRequest(time.Now().Unix()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request to update node")
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		err = parseResponseError(resp.Body)
		return errors.Wrapf(err, "failed to update node with twin id %d with status code %s", c.twinID, resp.Status)
	}
	defer resp.Body.Close()

	return
}

func (c RegistrarClient) reportUptime(report UptimeReport) (err error) {
	err = c.ensureNodeID()
	if err != nil {
		return err
	}

	url, err := url.JoinPath(c.baseURL, "nodes", fmt.Sprint(c.nodeID))
	if err != nil {
		return errors.Wrap(err, "failed to construct registrar url")
	}

	return
}

func (c RegistrarClient) getNode(id uint64) (node Node, err error) {
	url, err := url.JoinPath(c.baseURL, "nodes", fmt.Sprint(id))
	if err != nil {
		return node, errors.Wrap(err, "failed to construct registrar url")
	}

	return
}

func (c RegistrarClient) getNodeByTwinID(id uint64) (node Node, err error) {
	url, err := url.JoinPath(c.baseURL, "nodes")
	if err != nil {
		return node, errors.Wrap(err, "failed to construct registrar url")
	}

	return
}

func (c RegistrarClient) listNodes(opts ...ListNodeOpts) (nodes []Node, err error) {
	url, err := url.JoinPath(c.baseURL, "nodes")
	if err != nil {
		return nodes, errors.Wrap(err, "failed to construct registrar url")
	}
	return
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

func (c RegistrarClient) parseUpdateNodeOpts(node Node, opts []UpdateNodeOpts) Node {
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
