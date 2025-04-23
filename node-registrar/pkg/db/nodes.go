package db

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const ZOS4VersionKey = "zos_4"

// ListNodes retrieves all nodes from the database with applied filters and pagination
func (db *Database) ListNodes(filter NodeFilter, limit Limit) (nodes []Node, err error) {
	query := db.gormDB.Model(&Node{})

	if filter.NodeID != nil {
		query = query.Where("node_id= ?", *filter.NodeID)
	}
	if filter.FarmID != nil {
		query = query.Where("farm_id = ?", *filter.FarmID)
	}
	if filter.TwinID != nil {
		query = query.Where("twin_id = ?", *filter.TwinID)
	}

	// Filter by online status (node sent an uptime report in the last 30 minutes)
	if filter.Online != nil {
		// Calculate the cutoff time (30 minutes ago by default)
		cutoffMinutes := int64(30) // Default to 30 minutes
		if filter.LastSeen != nil {
			cutoffMinutes = *filter.LastSeen
		}
		cutoffTime := time.Now().Add(-time.Duration(cutoffMinutes) * time.Minute)

		if *filter.Online {
			// Online nodes: last_seen is not null and more recent than cutoff time
			query = query.Where("last_seen IS NOT NULL AND last_seen > ?", cutoffTime)
		} else {
			// Offline nodes: last_seen is null or older than cutoff time
			query = query.Where("last_seen IS NULL OR last_seen <= ?", cutoffTime)
		}
	} else if filter.LastSeen != nil {
		// If only LastSeen is provided without Online flag, show nodes active within that period
		cutoffTime := time.Now().Add(-time.Duration(*filter.LastSeen) * time.Minute)
		query = query.Where("last_seen IS NOT NULL AND last_seen > ?", cutoffTime)
	}

	offset := (limit.Page - 1) * limit.Size
	query = query.Offset(int(offset)).Limit(int(limit.Size))

	if result := query.Find(&nodes); result.Error != nil {
		return nil, result.Error
	}

	return nodes, nil
}

// GetNode retrieves a specific node by its nodeID
func (db *Database) GetNode(nodeID uint64) (node Node, err error) {
	if result := db.gormDB.First(&node, nodeID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return node, ErrRecordNotFound
		}
		return node, result.Error
	}
	return node, nil
}

// RegisterNode registers a new node in the database
func (db *Database) RegisterNode(node Node) (uint64, error) {
	if result := db.gormDB.Create(&node); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return 0, ErrRecordAlreadyExists
		}
		return 0, result.Error
	}
	return node.NodeID, nil
}

func (db *Database) UpdateNode(nodeID uint64, node Node) error {
	result := db.gormDB.Model(&Node{}).
		Where("node_id = ?", nodeID).
		Updates(node)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Uptime updates the uptime for a specific node
func (db *Database) GetUptimeReports(nodeID uint64, start, end time.Time) ([]UptimeReport, error) {
	var reports []UptimeReport
	result := db.gormDB.Where("node_id = ? AND timestamp BETWEEN ? AND ?",
		nodeID, start, end).Order("timestamp asc").Find(&reports)
	return reports, result.Error
}

func (db *Database) CreateUptimeReport(report *UptimeReport) error {
	// Start a transaction
	tx := db.gormDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Create the uptime report
	if err := tx.Create(report).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update the node's LastSeen field
	now := time.Now()
	if err := tx.Model(&Node{}).Where("node_id = ?", report.NodeID).Update("last_seen", now).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

func (db *Database) SetZOSVersion(version string) error {
	var current ZosVersion
	result := db.gormDB.Where(ZosVersion{Key: ZOS4VersionKey}).Attrs(ZosVersion{Version: version}).FirstOrCreate(&current)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		if current.Version == version {
			return errors.New("version already set")
		}
		return db.gormDB.Model(&current).
			Select("version").
			Update("version", version).Error
	}
	return nil
}

func (db *Database) GetZOSVersion() (string, error) {
	var setting ZosVersion
	if err := db.gormDB.Where("key = ?", "zos_4").First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrRecordNotFound
		}
		return "", err
	}
	return setting.Version, nil
}
