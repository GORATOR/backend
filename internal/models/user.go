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

type CreatedByUser interface {
	SetUserId(userId uint)
}

type CreatedByUserStruct struct {
	CreatedBy uint
}

func (cbu *CreatedByUserStruct) SetUserId(userId uint) {
	cbu.CreatedBy = userId
}

type User struct {
	gorm.Model
	CreatedBy     uint
	Username      string `gorm:"index:idx_username,unique"`
	Password      string
	Email         string `gorm:"index:idx_email,unique"`
	Avatar        string
	Active        bool
	Teams         []*Team         `gorm:"many2many:team_users;"`
	Organizations []*Organization `gorm:"many2many:org_users;"`
	Projects      []*Project      `gorm:"foreignKey:CreatedBy"`
}

func (u *User) CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	return createModel[User](data, tx)
}

func (u *User) UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	return updateModel[User](data, tx)
}

func (u *User) SetUserId(userId uint) {
	u.CreatedBy = userId
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

func (u *User) OnCreateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (u *User) OnReadParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (u *User) OnUpdateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}
