package models

import (
	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

const (
	UserEntityName = "user"
)

var (
	UserSelectFields = []string{
		"ID",
		"CreatedAt",
		"UpdatedAt",
		"Username",
		"Email",
		"Avatar",
		"Active",
	}
)

type User struct {
	gorm.Model
	Username      string `gorm:"index:idx_username,unique"`
	Password      string
	Email         string `gorm:"index:idx_email,unique"`
	Avatar        string
	Active        bool
	Teams         []*Team         `gorm:"many2many:team_users;"`
	Organizations []*Organization `gorm:"many2many:org_users;"`
}

func (u *User) CreateHashedPassword(plaintextPassword string, salt string) {
	u.Password = utils.HashPassword(plaintextPassword, salt)
}
