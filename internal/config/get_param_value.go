package config

import (
	"github.com/GORATOR/backend/internal/utils"
)

const (
	EnvAuthUserId = "GORATOR_AUTH_USER_ID"
)

func GetParamValue(param string) string {
	return utils.StringFromEnv(param, "")
}
