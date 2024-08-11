package config

import (
	"github.com/GORATOR/backend/internal/utils"
)

func IsDebug() bool {
	isDebug := utils.StringFromEnv("GORATOR_IS_DEBUG", "0")
	return isDebug == "1"
}
