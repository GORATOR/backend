package models

import "gorm.io/gorm"

type Team struct {
	gorm.Model
	Name          string
	Avatar        string
	Users         []*User         `gorm:"many2many:team_users;"`
	Organizations []*Organization `gorm:"many2many:org_teams;"`
}
