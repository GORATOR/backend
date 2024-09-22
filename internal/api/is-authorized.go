package api

import "net/http"

const (
	MessageUnauthorized = "Unauthorized"
)

func IsAuthorized(r *http.Request) (bool, int) {
	sessionID, err := r.Cookie("session")
	if err != nil {
		return false, 0
	}
	session, err := sessionStore.GetSession(sessionID.Value)
	return session != nil && err == nil, session.UserId
}
