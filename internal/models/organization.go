package models

import "gorm.io/gorm"

type Organization struct {
	gorm.Model
	Name   string
	Avatar string
	Active bool
	Teams  []*Team `gorm:"many2many:org_teams;"`
	Users  []*User `gorm:"many2many:org_users;"`
}
