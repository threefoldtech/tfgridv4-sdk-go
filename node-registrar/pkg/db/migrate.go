package db

import (
	"strings"

	"github.com/pkg/errors"
)

func (db Database) autoMigrate() error {
	if err := db.migrateNodes(); err != nil {
		return err
	}

	if err := db.gormDB.AutoMigrate(
		&Account{},
		&Farm{},
		&Node{},
		&UptimeReport{},
		&ZosVersion{},
	); err != nil {
		return errors.Wrap(err, "failed to migrate tables")
	}
	return nil
}

func (db Database) migrateNodes() error {
	// if nodes are already migrated skip migration
	if result := db.gormDB.First(&Node{}); result.Error == nil {
		return nil
	}

	type oldInterface struct {
		Name string `json:"name"`
		Mac  string `json:"mac"`
		IPs  string `json:"ips"`
	}

	type nodeType struct {
		NodeID uint64 `json:"node_id" gorm:"primaryKey;autoIncrement"`
		FarmID uint64 `json:"farm_id" gorm:"not null;check:farm_id> 0;foreignKey:FarmID;references:FarmID;constraint:OnDelete:RESTRICT"`
		TwinID uint64 `json:"twin_id" gorm:"not null;unique;check:twin_id > 0;foreignKey:TwinID;references:TwinID;constraint:OnDelete:RESTRICT"`

		Location Location `json:"location" gorm:"not null;type:json;serializer:json"`

		Resources    Resources      `json:"resources" gorm:"not null;type:json;serializer:json"`
		Interfaces   []oldInterface `gorm:"not null;type:json;serializer:json"`
		SecureBoot   bool           `json:"secure_boot"`
		Virtualized  bool           `json:"virtualized"`
		SerialNumber string         `json:"serial_number"`

		Approved bool `json:"approved"`
	}

	var nodes []nodeType
	result := db.gormDB.Model(&Node{}).Find(&nodes)
	if result.Error != nil {
		return result.Error
	}

	for _, node := range nodes {
		var interfaces []Interface
		for _, i := range node.Interfaces {
			ips := strings.Split(i.IPs, "/")
			newInterface := Interface{
				Name: i.Name,
				Mac:  i.Mac,
				IPs:  ips,
			}
			interfaces = append(interfaces, newInterface)
		}

		updatedNode := Node{
			NodeID:      node.NodeID,
			FarmID:      node.FarmID,
			TwinID:      node.TwinID,
			Location:    node.Location,
			Resources:   node.Resources,
			Interfaces:  interfaces,
			SecureBoot:  node.SecureBoot,
			Virtualized: node.Virtualized,
			Approved:    node.Approved,
		}

		err := db.gormDB.Model(&Node{}).Where("node_id = ?", node.NodeID).Updates(updatedNode).Error
		if err != nil {
			return err
		}
	}
	return nil
}
