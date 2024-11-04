package models

type Model interface {
	User | Team | Organization | Project
}

//type EntityInterface interface

type ModelCountResponse struct {
	Entity string `json:"entity"`
	Count  int64  `json:"count"`
}
