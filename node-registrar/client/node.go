package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
func (c *RegistrarClient) UpdateNode(updateOpts NodeUpdate) (err error) {
	return c.updateNode(updateOpts)
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
func (c *RegistrarClient) ListNodes(opts NodeFilter) (nodes []Node, err error) {
	return c.listNodesWithFilter(opts)
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
			return nodeID, errors.Wrap(err, "failed to send request to registrar the node")
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

// NodeUpdate represents update options for a node
type NodeUpdate struct {
	FarmID       *uint64
	Location     *Location
	Resources    *Resources
	Interfaces   []Interface
	SecureBoot   *bool
	Virtualized  *bool
	SerialNumber *string
	Status       *string
	Healthy      *bool
	Approved     *bool
}

func (c *RegistrarClient) updateNode(opts NodeUpdate) (err error) {
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
			return errors.Wrap(err, "failed here")
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
	nodes, err := c.ListNodes(NodeFilter{TwinID: &id})
	if err != nil {
		return
	}

	if len(nodes) == 0 {
		return node, ErrorNodeNotFound
	}

	return nodes[0], nil
}

// NodeFilter represents filtering options for listing nodes
type NodeFilter struct {
	NodeID   *uint64
	FarmID   *uint64
	TwinID   *uint64
	Status   *string
	Healthy  *bool
	Online   *bool
	LastSeen *int64
	Page     *uint32
	Size     *uint32
}

func (c *RegistrarClient) listNodesWithFilter(filter NodeFilter) (nodes []Node, err error) {
	url, err := url.JoinPath(c.baseURL, "nodes")
	if err != nil {
		return nodes, errors.Wrap(err, "failed to construct registrar url")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nodes, errors.Wrap(err, "failed to construct http request to the registrar")
	}

	q := req.URL.Query()
	data := parseListNodeOpts(filter)

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

	// try old interface format
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
			return nodes, errors.Wrap(err, "failed to get nodes with old interface format")
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

func (c *RegistrarClient) parseUpdateNodeOpts(node Node, update NodeUpdate) Node {
	if update.FarmID != nil {
		node.FarmID = *update.FarmID
	}
	if update.Location != nil {
		node.Location = *update.Location
	}
	if update.Resources != nil {
		node.Resources = *update.Resources
	}
	if len(update.Interfaces) > 0 {
		node.Interfaces = update.Interfaces
	}
	if update.SerialNumber != nil {
		node.SerialNumber = *update.SerialNumber
	}
	if update.SecureBoot != nil {
		node.SecureBoot = *update.SecureBoot
	}
	if update.Virtualized != nil {
		node.Virtualized = *update.Virtualized
	}
	if update.Status != nil {
		node.Virtualized = *update.Virtualized
	}
	if update.Healthy != nil {
		node.Virtualized = *update.Virtualized
	}
	if update.Approved != nil {
		node.Virtualized = *update.Virtualized
	}

	return node
}

func parseListNodeOpts(filter NodeFilter) map[string]any {
	data := map[string]any{}

	if filter.NodeID != nil {
		data["node_id"] = *filter.NodeID
	}
	if filter.TwinID != nil {
		data["twin_id"] = *filter.TwinID
	}
	if filter.FarmID != nil {
		data["farm_id"] = *filter.FarmID
	}
	if filter.Status != nil && *filter.Status != "" {
		data["status"] = *filter.Status
	}
	if filter.Healthy != nil {
		data["healthy"] = *filter.Healthy
	}
	if filter.Online != nil {
		data["online"] = *filter.Online
	}
	if filter.LastSeen != nil {
		data["last_seen"] = *filter.LastSeen
	}

	page := uint32(1)
	if filter.Page != nil {
		page = *filter.Page
	}
	data["page"] = page

	size := uint32(DefaultPageSize)
	if filter.Size != nil {
		size = *filter.Size
	}

	data["size"] = size

	return data
}

type oldInterfaceFormat struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
	IPs  string `json:"ips"`
}

type oldFormatNodeType struct {
	Interfaces []oldInterfaceFormat `json:"interfaces"`

	NodeID       uint64     `json:"node_id"`
	FarmID       uint64     `json:"farm_id"`
	TwinID       uint64     `json:"twin_id"`
	Location     Location   `json:"location"`
	Resources    Resources  `json:"resources"`
	SecureBoot   bool       `json:"secure_boot"`
	Virtualized  bool       `json:"virtualized"`
	SerialNumber string     `json:"serial_number"`
	LastSeen     *time.Time `json:"last_seen"`
	Online       bool       `json:"online"`
	Approved     bool
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

	oldFormatNode := oldFormatNodeType{
		Interfaces: oldInterfaces,

		NodeID:       node.NodeID,
		FarmID:       node.FarmID,
		TwinID:       node.TwinID,
		Location:     node.Location,
		Resources:    node.Resources,
		SecureBoot:   node.SecureBoot,
		Virtualized:  node.Virtualized,
		SerialNumber: node.SerialNumber,
		LastSeen:     node.LastSeen,
		Online:       node.Online,
		Approved:     node.Approved,
	}

	err = json.NewEncoder(&body).Encode(oldFormatNode)
	if err != nil {
		return
	}

	return
}

func parseResponseBodyToNewInterfaceFormat(nodeBytes []byte) (Node, error) {
	var oldFormatNode oldFormatNodeType
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
	return Node{
		Interfaces: newFormat,

		NodeID:       oldFormatNode.NodeID,
		FarmID:       oldFormatNode.FarmID,
		TwinID:       oldFormatNode.TwinID,
		Location:     oldFormatNode.Location,
		Resources:    oldFormatNode.Resources,
		SecureBoot:   oldFormatNode.SecureBoot,
		Virtualized:  oldFormatNode.Virtualized,
		SerialNumber: oldFormatNode.SerialNumber,
		LastSeen:     oldFormatNode.LastSeen,
		Online:       oldFormatNode.Online,
		Approved:     oldFormatNode.Approved,
	}, nil
}
