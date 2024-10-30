package models

type Entity interface {
	User | Team | Organization
}

type EntityCountResponse struct {
	Entity string `json:"entity"`
	Count  int64  `json:"count"`
}
