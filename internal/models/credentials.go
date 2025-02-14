package models

type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CredentialsResponse struct {
	User      *UserResponse `json:"user"`
	SessionId string        `json:"sessionId"`
}
