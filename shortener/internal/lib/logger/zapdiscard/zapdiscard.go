package zapdiscard

import (
	"go.uber.org/zap"
)

func New() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}
