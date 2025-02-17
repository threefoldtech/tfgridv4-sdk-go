package client

import "fmt"

var ErrorNodeNotFround = fmt.Errorf("failed to get requested node from node regiatrar")

func (c RegistrarClient) RegisterNode(
	farmID uint64,
	twinID uint64,
	interfaces Interface,
	location Location,
	resources Resources,
	serialNumber string,
	secureBoot,
	virtualized bool,
) (node Node, err error) {
	return
}

func (c RegistrarClient) UpdateNode(
	nodeID uint64,
	farmID uint64,
	interfaces Interface,
	location Location,
	resources Resources,
	serialNumber string,
	secureBoot,
	virtualized bool,
) (err error) {
	return
}

func (c RegistrarClient) ReportUptime(id uint64, report UptimeReport) (err error) {
	return
}

func (c RegistrarClient) GetNode(id uint64) (node Node, err error) {
	return
}

func (c RegistrarClient) GetNodeByTwinID(id uint64) (node Node, err error) {
	return
}

func (c RegistrarClient) ListNodes(opts ...NodeOpts) (nodes []Node, err error) {
	return
}

type nodeCfg struct {
	nodeID  uint64
	farmID  uint64
	twinID  uint64
	status  string
	healthy bool
	page    uint32
	size    uint32
}

type NodeOpts func(*nodeCfg)

func NodeWithNodeID(id uint64) NodeOpts {
	return func(n *nodeCfg) {
		n.nodeID = id
	}
}

func NodeWithFarmID(id uint64) NodeOpts {
	return func(n *nodeCfg) {
		n.farmID = id
	}
}

func NodeWithStatus(status string) NodeOpts {
	return func(n *nodeCfg) {
		n.status = status
	}
}

func NodeHealthy() NodeOpts {
	return func(n *nodeCfg) {
		n.healthy = true
	}
}

func NodeWithTwinID(id uint64) NodeOpts {
	return func(n *nodeCfg) {
		n.twinID = id
	}
}

func NodeWithPage(page uint32) NodeOpts {
	return func(n *nodeCfg) {
		n.page = page
	}
}

func NodeWithSize(size uint32) NodeOpts {
	return func(n *nodeCfg) {
		n.size = size
	}
}
