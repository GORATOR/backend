package models

import (
	"net/http"

	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

const (
	UserModelName = "user"
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

func (u *User) GetName() string {
	return UserModelName
}

func (u *User) ParseQueryString(endpoint string, query *gorm.DB, r *http.Request) {
	parseUsersQuery(query, r)
}

func (u *User) GetSelectFields() *[]string {
	return &UserSelectFields
}

func (u *User) FindAll(query *gorm.DB) (interface{}, error) {
	users, err := findAll[User](*u.GetSelectFields(), query)
	return users, err
}
