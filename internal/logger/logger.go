package logger

import (
	"github.com/jeanmolossi/verbose-adventure/internal/config"
	"go.uber.org/zap"
)

func NewZap(cfg *config.Config) (*zap.Logger, error) {
	zcfg := zap.NewProductionConfig()

	_ = zcfg.Level.UnmarshalText([]byte(cfg.LogLevel))

	return zcfg.Build()
}
