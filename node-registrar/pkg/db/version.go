package db

import (
	"errors"

	"gorm.io/gorm"
)

func (db *Database) SetZOSVersion(version string) error {
	var current ZosVersion
	result := db.gormDB.Where(ZosVersion{Key: ZOS4VersionKey}).Attrs(ZosVersion{Version: version}).FirstOrCreate(&current)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		if current.Version == version {
			return ErrVersionAlreadySet
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
