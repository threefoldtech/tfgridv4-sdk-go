package db

import (
	"errors"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Account struct {
	TwinID    uint64         `gorm:"primaryKey;autoIncrement" json:"twin_id"`
	Relays    pq.StringArray `gorm:"type:text[];default:'{}'" json:"relays" swaggertype:"array,string"` // Optional list of relay domains
	RMBEncKey string         `gorm:"type:text" json:"rmb_enc_key"`                                      // Optional base64 encoded public key for rmb communication
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	// The public key (ED25519 for nodes, ED25519 or SR25519 for farmers) in the more standard base64 since we are moving from substrate echo system?
	// (still SS58 can be used or plain base58 ,TBD)
	PublicKey string `gorm:"type:text;not null;unique" json:"public_key"`
	// Relations | likely we need to use OnDelete:RESTRICT (Prevent Twin deletion if farms exist)
	// @swagger:ignore
	Farms []Farm `gorm:"foreignKey:TwinID;references:TwinID;constraint:OnDelete:RESTRICT"`
}

type Farm struct {
	FarmID         uint64    `gorm:"primaryKey;autoIncrement" json:"farm_id"`
	FarmName       string    `gorm:"size:40;not null;unique;check:farm_name <> ''" json:"farm_name" binding:"alphanum,required"`
	TwinID         uint64    `json:"twin_id" binding:"required" gorm:"not null;check:twin_id > 0"` // Farmer account reference
	StellarAddress string    `json:"stellar_address" binding:"required,startswith=G,len=56,alphanum,uppercase"`
	Dedicated      bool      `json:"dedicated"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// @swagger:ignore
	Nodes []Node `gorm:"foreignKey:FarmID;references:FarmID;constraint:OnDelete:RESTRICT" json:"nodes"`
}

type Node struct {
	NodeID uint64 `json:"node_id" gorm:"primaryKey;autoIncrement"`
	// Constraints set to prevents unintended account deletion if linked Farms/nodes exist.
	FarmID uint64 `json:"farm_id" gorm:"not null;check:farm_id> 0;foreignKey:FarmID;references:FarmID;constraint:OnDelete:RESTRICT"`
	TwinID uint64 `json:"twin_id" gorm:"not null;unique;check:twin_id > 0;foreignKey:TwinID;references:TwinID;constraint:OnDelete:RESTRICT"` // Node account reference

	Location Location `json:"location" gorm:"not null;type:json;serializer:json"`

	// PublicConfig PublicConfig `json:"public_config" gorm:"type:json"`
	Resources    Resources   `json:"resources" gorm:"not null;type:json;serializer:json"`
	Interfaces   []Interface `gorm:"not null;type:json;serializer:json"`
	SecureBoot   bool        `json:"secure_boot"`
	Virtualized  bool        `json:"virtualized"`
	SerialNumber string      `json:"serial_number"`

	UptimeReports []UptimeReport `json:"uptime" gorm:"foreignKey:NodeID;references:NodeID;constraint:OnDelete:CASCADE"`
	LastSeen      time.Time      `json:"last_seen" gorm:"index"` // Last time the node sent Uptime report
	Online        bool           `json:"online" gorm:"-"`        // Computed field, not stored in database
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Approved      bool           `json:"approved"`
}

func (n *Node) BeforeCreate(tx *gorm.DB) (err error) {
	if len(n.Interfaces) == 0 {
		return errors.New("interfaces must not be empty")
	}
	return nil
}

type UptimeReport struct {
	ID         uint64        `gorm:"primaryKey;autoIncrement"`
	NodeID     uint64        `gorm:"index" json:"node_id"`
	Duration   time.Duration `json:"duration" swaggertype:"integer"` // Uptime duration for this period
	Timestamp  time.Time     `json:"timestamp" gorm:"index"`
	WasRestart bool          `json:"was_restart"` // True if this report followed a restart
	CreatedAt  time.Time     `json:"created_at"`
}

type ZosVersion struct {
	Key     string `gorm:"primaryKey;size:50"`
	Version string `gorm:"not null"`
}

type Interface struct {
	Name string   `json:"name"`
	Mac  string   `json:"mac"`
	IPs  []string `json:"ips" gorm:"not null,type:text[];default:'{}'"`
}

type Resources struct {
	HRU uint64 `json:"hru"`
	SRU uint64 `json:"sru"`
	CRU uint64 `json:"cru"`
	MRU uint64 `json:"mru"`
}

type Location struct {
	Country   string `json:"country" gorm:"not null"`
	City      string `json:"city" gorm:"not null"`
	Longitude string `json:"longitude" gorm:"not null"`
	Latitude  string `json:"latitude" gorm:"not null"`
}

type NodeFilter struct {
	NodeID   *uint64 `form:"node_id"`
	FarmID   *uint64 `form:"farm_id"`
	TwinID   *uint64 `form:"twin_id"`
	Status   string  `form:"status"`
	Healthy  bool    `form:"healthy"`
	Online   *bool   `form:"online"`    // Filter by online status (true = online, false = offline, nil = both)
	LastSeen *int64  `form:"last_seen"` // Filter nodes last seen within this many minutes
}

type FarmFilter struct {
	FarmName *string `form:"farm_name"`
	FarmID   *uint64 `form:"farm_id"`
	TwinID   *uint64 `form:"twin_id"`
}

// Limit used for pagination
type Limit struct {
	Size uint64 `form:"size"`
	Page uint64 `form:"page"`
}

// DefaultLimit returns the default values for the pagination
func DefaultLimit() Limit {
	return Limit{
		Size: 50,
		Page: 1,
	}
}
