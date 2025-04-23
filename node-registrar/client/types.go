package client

import (
	"time"
)

type Account struct {
	TwinID    uint64   `json:"twin_id"`
	Relays    []string `json:"relays"`      // Optional list of relay domains
	RMBEncKey string   `json:"rmb_enc_key"` // Optional base64 encoded public key for rmb communication
	PublicKey string   `json:"public_key"`
}

type Farm struct {
	FarmID         uint64 `json:"farm_id"`
	FarmName       string `json:"farm_name"`
	TwinID         uint64 `json:"twin_id"`
	Dedicated      bool   `json:"dedicated"`
	StellarAddress string `json:"stellar_address"`
}

type Node struct {
	NodeID        uint64         `json:"node_id"`
	FarmID        uint64         `json:"farm_id"`
	TwinID        uint64         `json:"twin_id"`
	Location      Location       `json:"location"`
	Resources     Resources      `json:"resources"`
	Interfaces    []Interface    `json:"interfaces"`
	SecureBoot    bool           `json:"secure_boot"`
	Virtualized   bool           `json:"virtualized"`
	SerialNumber  string         `json:"serial_number"`
	UptimeReports []UptimeReport `json:"uptime"`
	LastSeen      *time.Time     `json:"last_seen"`
	Online        bool           `json:"online"`
	Approved      bool
}

type UptimeReport struct {
	Uptime    time.Duration `json:"uptime"`
	Timestamp time.Time     `json:"timestamp"`
}

type ZosVersion struct {
	Version       string `json:"version"`
	SafeToUpgrade bool   `json:"safe_to_upgrade"`
}

type Interface struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
	IPs  string `json:"ips"`
}

type Resources struct {
	HRU uint64 `json:"hru"`
	SRU uint64 `json:"sru"`
	CRU uint64 `json:"cru"`
	MRU uint64 `json:"mru"`
}

type Location struct {
	Country   string `json:"country"`
	City      string `json:"city"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}
