package models

import "gorm.io/gorm"

const (
	OrganizationModelName = "organization"
)

type Organization struct {
	gorm.Model
	Name   string
	Avatar string
	Active bool
	Teams  []*Team `gorm:"many2many:org_teams;"`
	Users  []*User `gorm:"many2many:org_users;"`
}
