package models

import (
	"encoding/json"
	"fmt"
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

type UserInput struct {
	ID              uint
	Username        string
	Password        string
	Email           string
	Avatar          string
	TeamIds         []uint
	OrganizationIds []uint
	ProjectIds      []uint
}

func (u *User) CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input UserInput
	var user User
	user.CreatedBy = userId
	err := json.Unmarshal(data, &input)
	if err != nil {
		fmt.Print("CreateModel json.Unmarshal error ", err)
		return nil, err
	}
	user.Active = true
	user.Avatar = input.Avatar
	user.Username = input.Username
	user.Email = input.Email
	user.SetPassword(input.Password)

	insertError := tx.Transaction(func(tx *gorm.DB) error {
		insertResult := tx.Save(&user)
		if insertResult.Error != nil {
			fmt.Print("CreateModel tx.Save error ", insertResult.Error)
			return insertResult.Error
		}

		if len(input.TeamIds) > 0 {
			var teams []*Team
			tx.Model(&user).Association(
				bindRelatedModelsToModel(
					tx,
					UserModelName,
					TeamModelName,
					user.ID,
					input.TeamIds,
				),
			).Find(&teams)
			user.Teams = teams
		}

		if len(input.OrganizationIds) > 0 {
			var orgs []*Organization
			tx.Model(&user).Association(
				bindRelatedModelsToModel(
					tx,
					UserModelName,
					OrganizationModelName,
					user.ID,
					input.OrganizationIds,
				),
			).Find(&orgs)
			user.Organizations = orgs
		}

		return nil
	})

	user.Password = ""

	return user, insertError
}

func (u *User) UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input UserInput
	var user User
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, err
	}
	user.SetPassword(input.Password)
	updates := map[string]interface{}{
		"username": input.Username,
		"avatar":   input.Avatar,
		"password": user.Password,
	}
	updateError := tx.Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&User{}).Where("active = ? and id = ?", true, input.ID).Updates(updates)
		if updateResult.Error != nil {
			fmt.Printf("update user with id %d error %s", input.ID, updateResult.Error)
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			fmt.Printf("update user RowsAffected = 0")
			return fmt.Errorf("no active users with id = %d", input.ID)
		}

		findResult := tx.Model(&user).Where("active = ? and id = ?", true, input.ID).Find(&user)
		if findResult.Error != nil {
			fmt.Print("after updateModel (find) error ", findResult.Error)
			return findResult.Error
		}

		tx.Exec("DELETE FROM org_users WHERE user_id = ?", input.ID)
		if len(input.TeamIds) > 0 {
			var orgs []*Organization
			tx.Model(&user).Association(
				bindRelatedModelsToModel(
					tx,
					UserModelName,
					OrganizationModelName,
					user.ID,
					input.OrganizationIds,
				),
			).Find(&orgs)
			user.Organizations = orgs
		}

		tx.Exec("DELETE FROM team_users WHERE user_id = ?", input.ID)
		if len(input.OrganizationIds) > 0 {
			var teams []*Team
			tx.Model(&user).Association(
				bindRelatedModelsToModel(
					tx,
					UserModelName,
					TeamModelName,
					user.ID,
					input.TeamIds,
				),
			).Find(&teams)
			user.Teams = teams
		}
		return nil
	})

	if updateError != nil {
		return nil, updateError
	}

	user.Password = ""

	return user, updateError
}

func (u *User) SetUserId(userId uint) {
	u.CreatedBy = userId
}

func (u *User) SetPassword(plainText string) {
	salt := utils.StringFromEnv("GORATOR_SALT", "")
	if salt == "" {
		fmt.Printf("Empty salt!")
	}
	u.CreateHashedPassword(plainText, salt)
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

func (u *User) ReadById(db *gorm.DB, id uint) (interface{}, error) {
	return readById(db, id, u)
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
