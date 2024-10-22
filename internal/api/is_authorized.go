package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/GORATOR/backend/internal/config"
)

const (
	MessageUnauthorized = "Unauthorized"
	sessionEmpty        = ""
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
	sessionID := tryGetSessionId(r)
	if sessionID == sessionEmpty {
		return false, 0
	}
	session, err := sessionStore.GetSession(sessionID)
	return session != nil && err == nil, session.UserId
}

func tryGetSessionId(r *http.Request) string {
	sessionID, errSession := r.Cookie("session")
	headerValue := r.Header.Get("X-Session-Id")
	if errSession == nil && sessionID.Value != sessionEmpty {
		return sessionID.Value
	}
	return headerValue
}
