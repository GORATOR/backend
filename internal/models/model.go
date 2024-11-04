package models

type Model interface {
	GetName() string
}

type ModelCountResponse struct {
	Entity string `json:"entity"`
	Count  int64  `json:"count"`
}
