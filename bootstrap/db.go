package bootstrap

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tdl-filegram/internal/model"
)

// NewDB 初始化 SQLite（纯 Go 驱动，免 CGO），并自动迁移任务表
func NewDB(path string, log *zap.Logger) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.Job{}); err != nil {
		return nil, err
	}
	model.DB = db
	log.Info("sqlite ready", zap.String("path", path))
	return db, nil
}
