package models

type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CredentialsResponse struct {
	User      User   `json:"user"`
	SessionId string `json:"session_id"`
}
