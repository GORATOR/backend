package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username      string
	Password      string
	Email         string
	Teams         []*Team         `gorm:"many2many:team_users;"`
	Organizations []*Organization `gorm:"many2many:org_users;"`
}
