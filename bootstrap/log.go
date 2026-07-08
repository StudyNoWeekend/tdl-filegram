package bootstrap

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger 初始化 zap 结构化日志，同时输出到文件与控制台
func NewLogger(cfg LogConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if cfg.Level == "debug" {
		level = zapcore.DebugLevel
	}
	if cfg.Dir == "" {
		cfg.Dir = "logs"
	}
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, err
	}
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.Dir, "app.log"),
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     7,
	})
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), fileWriter, level),
		zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.AddSync(os.Stdout), level),
	)
	return zap.New(core), nil
}
