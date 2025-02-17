package client

import "fmt"

var ErrorFarmNotFround = fmt.Errorf("failed to get requested farm from node regiatrar")

func (c RegistrarClient) CreateFarm(farmName string, twinID uint64, dedicated bool) (farm Farm, err error) {
	return
}

func (c RegistrarClient) UpdateFarm(farmID uint64, farmName string) (err error) {
	return
}

func (c RegistrarClient) GetFarm(id uint64) (farm Farm, err error) {
	return
}

func (c RegistrarClient) ListFarms(opts ...FarmOpts) (farms []Farm, err error) {
	return
}

type farmCfg struct {
	farmName string
	farmID   uint64
	twinID   uint64
	page     uint32
	size     uint32
}

type FarmOpts func(*farmCfg)

func FarmWithName(name string) FarmOpts {
	return func(n *farmCfg) {
		n.farmName = name
	}
}

func FarmWithFarmID(id uint64) FarmOpts {
	return func(n *farmCfg) {
		n.farmID = id
	}
}

func FarmWithTwinID(id uint64) FarmOpts {
	return func(n *farmCfg) {
		n.twinID = id
	}
}

func FarmWithPage(page uint32) FarmOpts {
	return func(n *farmCfg) {
		n.page = page
	}
}

func FarmWithSize(size uint32) FarmOpts {
	return func(n *farmCfg) {
		n.size = size
	}
}
