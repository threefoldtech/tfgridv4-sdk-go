package db

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (db Database) autoMigrate() error {
	if err := db.gormDB.AutoMigrate(
		&Account{},
		&Farm{},
		&Node{},
		&UptimeReport{},
		&ZosVersion{},
	); err != nil {
		return errors.Wrap(err, "failed to migrate tables")
	}

	if err := db.migrateNodes(); err != nil {
		return err
	}

	err := db.MigrateNodeLastSeen()
	if err != nil {
		return errors.Wrap(err, "failed to migrate node last seen data")
	}

	return nil
}

func (db Database) migrateNodes() error {
	// if nodes are already migrated skip migration
	if result := db.gormDB.First(&Node{}); result.Error == nil {
		log.Info().Msg("nodes Interfaces are already migrated")
		return nil
	}

	type oldInterface struct {
		Name string `json:"name"`
		Mac  string `json:"mac"`
		IPs  string `json:"ips"`
	}

	//  we'd only load the data we actually need from the database
	type nodeType struct {
		NodeID     uint64         `json:"node_id" gorm:"primaryKey"`
		Interfaces []oldInterface `gorm:"not null;type:json;serializer:json"`
	}

	// Use a single transaction for all updates to ensure atomicity
	return db.Transaction(func(tx *gorm.DB) error {
		var nodes []nodeType
		if err := tx.Model(&Node{}).Find(&nodes).Error; err != nil {
			return err
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

			// skip the node if it has no interfaces
			if len(interfaces) != 0 {
				// Update only the interfaces field
				if err := tx.Model(&Node{}).
					Where("node_id = ?", node.NodeID).
					Update("interfaces", interfaces).Error; err != nil {
					return err // This will roll back the entire transaction
				}

				log.Info().Uint64("node_id", node.NodeID).Msg("Migration: updating node")
			}

		}

		return nil
	})
}

// MigrateNodeLastSeen updates the LastSeen field for existing nodes that don't have it set
func (db Database) MigrateNodeLastSeen() error {
	query := `
        UPDATE nodes n
        SET last_seen = (
            SELECT MAX(timestamp)
            FROM uptime_reports ur
            WHERE ur.node_id = n.node_id
        )
        WHERE (last_seen IS NULL OR last_seen = '0001-01-01 00:00:00+00')
        AND EXISTS (
            SELECT 1
            FROM uptime_reports ur
            WHERE ur.node_id = n.node_id
        )
    `

	result := db.gormDB.Exec(query)
	if result.Error == nil {
		log.Info().Msgf("Migration: Updated LastSeen for %d nodes", result.RowsAffected)
	}
	return result.Error
}
