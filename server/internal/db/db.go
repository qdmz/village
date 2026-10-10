// Package db 初始化数据库连接（SQLite / MySQL）。
package db

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/qdmz/village-mgmt/internal/config"
	"github.com/qdmz/village-mgmt/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	var dial gorm.Dialector
	switch config.C.Database.Driver {
	case "mysql", "mariadb":
		dial = mysql.Open(config.C.Database.DSN)
	case "sqlite", "":
		dsn := config.C.Database.DSN
		if dsn == "" {
			dsn = "./data/village.db"
		}
		if err := os.MkdirAll(filepath.Dir(dsn), 0o755); err != nil {
			return err
		}
		dial = sqlite.Open(dsn + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(0)")
	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.C.Database.Driver)
	}
	d, err := gorm.Open(dial, &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		return err
	}
	if sqlDB, err := d.DB(); err == nil {
		sqlDB.SetConnMaxLifetime(time.Hour)
		if config.C.Database.Driver == "sqlite" || config.C.Database.Driver == "" {
			sqlDB.SetMaxOpenConns(1)
		} else {
			sqlDB.SetMaxOpenConns(20)
		}
	}
	DB = d
	return d.AutoMigrate(model.All()...)
}

func GetSetting(key, def string) string {
	var s model.Setting
	if err := DB.Where("`key` = ?", key).First(&s).Error; err != nil {
		return def
	}
	return s.Value
}

func SetSetting(key, value string) error {
	return DB.Save(&model.Setting{Key: key, Value: value}).Error
}
