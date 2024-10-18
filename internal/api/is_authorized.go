package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/GORATOR/backend/internal/config"
)

const (
	MessageUnauthorized = "Unauthorized"
)

func IsAuthorized(r *http.Request) (bool, int) {
	if config.IsDebug() && config.IsParamSet(config.EnvSkipAuthCheck) {
		debugUserId, err := strconv.Atoi(config.GetParamValue(config.EnvAuthUserId))
		if err != nil {
			fmt.Printf("config.GetParamValue Atoi error: %s", err)
			return false, 0
		}
		return true, debugUserId
	}
	sessionID, err := r.Cookie("session")
	if err != nil {
		return false, 0
	}
	session, err := sessionStore.GetSession(sessionID.Value)
	return session != nil && err == nil, session.UserId
}
