package models

type Entity interface {
	User | Team | Organization | Project
}

//type EntityInterface interface

type EntityCountResponse struct {
	Entity string `json:"entity"`
	Count  int64  `json:"count"`
}
