package models

type Entity interface {
	User | Team | Organization
}
