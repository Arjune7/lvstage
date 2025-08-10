// internal/observability/logger.go
package observability

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

// InitializeZap initializes the global logger.
func InitializeZap() error {
	var err error
	Logger, err = zap.NewProduction() // or zap.NewDevelopment()
	if err != nil {
		return err
	}
	deferFunc := zap.RedirectStdLog(Logger) // redirect stdlib log

	// Prevent early GC of deferFunc
	_ = deferFunc

	return nil
}
