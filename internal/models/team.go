package models

import "gorm.io/gorm"

const (
	TeamEntityName = "team"
)

type Team struct {
	gorm.Model
	Name          string
	Avatar        string
	Active        bool
	Users         []*User         `gorm:"many2many:team_users;"`
	Organizations []*Organization `gorm:"many2many:org_teams;"`
}
