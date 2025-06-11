package db

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (db *Database) ListFarms(filter FarmFilter, limit Limit) (farms []Farm, err error) {
	query := db.gormDB.Model(&Farm{})

	if filter.FarmName != nil {
		query = query.Where("farm_name ILIKE ?", "%"+*filter.FarmName+"%") // Case-insensitive search
	}
	if filter.FarmID != nil {
		query = query.Where("farm_id = ?", *filter.FarmID)
	}
	if filter.TwinID != nil {
		query = query.Where("twin_id = ?", *filter.TwinID)
	}

	offset := (limit.Page - 1) * limit.Size
	query = query.Offset(int(offset)).Limit(int(limit.Size))

	if result := query.Find(&farms); result.Error != nil {
		return nil, result.Error
	}

	return farms, nil
}

func (db *Database) GetFarm(farmID uint64) (farm Farm, err error) {
	if result := db.gormDB.First(&farm, farmID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return farm, ErrRecordNotFound
		}
		return farm, result.Error
	}

	return
}

func (db *Database) CreateFarm(farm Farm) (uint64, error) {
	if err := db.gormDB.Create(&farm).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "23505") {
			return 0, ErrRecordAlreadyExists
		}
		return 0, err
	}

	return farm.FarmID, nil
}

func (db *Database) UpdateFarm(farmID uint64, name string, stellarAddr string) (err error) {
	update := map[string]interface{}{}

	if len(name) != 0 {
		update["farm_name"] = name
	}
	if len(stellarAddr) != 0 {
		update["stellar_address"] = stellarAddr
	}

	result := db.gormDB.Model(&Farm{}).Where("farm_id = ?", farmID).Updates(update)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// ApproveNodes approves multiple nodes for a specific farm
func (db *Database) ApproveNodes(farmID uint64, nodeIDs []uint64) error {
	// Start a transaction to ensure all updates are atomic
	tx := db.gormDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	result := tx.Model(&Node{}).
		Where("farm_id = ? AND node_id IN ? AND approved = ?", farmID, nodeIDs, false).
		Update("approved", true)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	// Check if all nodes were found and updated
	if int(result.RowsAffected) != len(nodeIDs) {
		tx.Rollback()
		return fmt.Errorf("some nodes were not found, do not belong to farm %d, or are already approved", farmID)
	}

	// Commit the transaction
	return tx.Commit().Error
}
