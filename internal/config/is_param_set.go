package config

import (
	"github.com/GORATOR/backend/internal/utils"
)

const (
	envPrefix        = "GORATOR_"
	EnvSkipAuthCheck = "GORATOR_SKIP_AUTH_CHECK"
)

func IsParamSet(param string) bool {
	var value = utils.StringFromEnv(param, "0")
	return value == "1"
}
