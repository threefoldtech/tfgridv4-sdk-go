package db

import (
	"errors"
	"reflect"

	"gorm.io/gorm"
)

func (db *Database) SetZOSVersion(version ZosVersion) error {
	var current ZosVersion
	result := db.gormDB.Where(ZosVersion{Key: ZOS4VersionKey}).Attrs(ZosVersion{Version: version.Version}).FirstOrCreate(&current)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		if reflect.DeepEqual(current, version) {
			return ErrVersionAlreadySet
		}
		return db.gormDB.Model(&current).
			Select("version").
			Updates(version).Error
	}
	return nil
}

func (db *Database) GetZOSVersion() (version ZosVersion, err error) {
	if err := db.gormDB.Where("key = ?", "zos_4").First(&version).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return version, ErrRecordNotFound
		}
		return version, err
	}
	return version, nil
}
