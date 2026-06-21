package db

import (
	"errors"

	"gorm.io/gorm"
)

const zos4Key = "zos_4"

func (db *Database) SetZOSVersion(version ZosVersion) error {
	var current ZosVersion
	result := db.gormDB.Where(ZosVersion{Key: ZOS4VersionKey}).Attrs(ZosVersion{Version: version.Version}).FirstOrCreate(&current)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		update := map[string]any{}
		if current.Version != version.Version {
			update["version"] = version.Version
		}
		if current.SafeToUpgrade != version.SafeToUpgrade {
			update["safe_to_upgrade"] = version.SafeToUpgrade
		}
		if len(update) == 0 {
			return ErrVersionAlreadySet
		}
		return db.gormDB.Model(&current).
			Where("key = ?", zos4Key).
			Updates(version).Error
	}
	return nil
}

func (db *Database) GetZOSVersion() (version ZosVersion, err error) {
	if err := db.gormDB.Where("key = ?", zos4Key).First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return version, ErrRecordNotFound
		}
		return version, err
	}
	return version, nil
}
