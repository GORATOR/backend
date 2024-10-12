package models

import "gorm.io/gorm"

const (
	UserEntityName = "user"
)

type User struct {
	gorm.Model
	Username      string
	Password      string
	Email         string
	Avatar        string
	Active        bool
	Teams         []*Team         `gorm:"many2many:team_users;"`
	Organizations []*Organization `gorm:"many2many:org_users;"`
}
