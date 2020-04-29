package log

import (
	"go.uber.org/zap"
	"testing"
)

func TestFactory_EnableLevel(t *testing.T) {
	factory := NewLogFactory("tmp", "unit_test", zap.DebugLevel)
	factory.Bg().Warn("Warn")
	factory.Bg().Info("Info")
	factory.Bg().Error("Error")
	factory.Bg().Debug("Debug")

	factory.EnableLevel(zap.InfoLevel)
	factory.Bg().Warn("Warn")
	factory.Bg().Info("Info")
	factory.Bg().Error("Error")
	factory.Bg().Debug("Debug")

	factory.EnableLevel(zap.ErrorLevel)
	factory.Bg().Warn("Warn")
	factory.Bg().Info("Info")
	factory.Bg().Error("Error")
	factory.Bg().Debug("Debug")

	factory.EnableLevel(zap.WarnLevel)
	factory.Bg().Warn("Warn")
	factory.Bg().Info("Info")
	factory.Bg().Error("Error")
	factory.Bg().Debug("Debug")
}
