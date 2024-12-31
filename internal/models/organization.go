package models

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

const (
	OrganizationModelName = "organization"
)

type Organization struct {
	gorm.Model
	CreatedByUserStruct
	Name   string
	Avatar string
	Active bool
	Teams  []*Team `gorm:"many2many:org_teams;"`
	Users  []*User `gorm:"many2many:org_users;"`
}

type OrganizationInput struct {
	ID      uint
	Name    string
	Avatar  string
	TeamIds []uint
	UserIds []uint
}

func (o *Organization) CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input OrganizationInput
	var org Organization
	org.SetUserId(userId)
	err := json.Unmarshal(data, &input)
	if err != nil {
		fmt.Print("CreateModel json.Unmarshal error ", err)
		return nil, err
	}
	org.Active = true
	org.Avatar = input.Avatar
	org.Name = input.Name

	insertError := tx.Transaction(func(tx *gorm.DB) error {
		insertResult := tx.Save(&org)
		if insertResult.Error != nil {
			fmt.Print("CreateModel tx.Save error ", insertResult.Error)
			return insertResult.Error
		}

		if len(input.TeamIds) > 0 {
			var teams []*Team
			tx.Model(&org).Association(
				bindModelToRelatedModels(
					tx,
					OrganizationModelName,
					TeamModelName,
					org.ID,
					input.TeamIds,
				),
			).Find(&teams)
			org.Teams = teams
		}

		if len(input.UserIds) > 0 {
			var users []*User
			tx.Model(&org).Omit("password").Association(
				bindModelToRelatedModels(
					tx,
					OrganizationModelName,
					UserModelName,
					org.ID,
					input.UserIds,
				),
			).Find(&users)
			org.Users = users
		}

		return nil
	})

	return org, insertError
}

func (o *Organization) UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input OrganizationInput
	var org Organization
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"name": input.Name, "avatar": input.Avatar}
	updateError := tx.Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&Organization{}).Where("active = ? and id = ?", true, input.ID).Updates(updates)
		if updateResult.Error != nil {
			fmt.Printf("update organization with id %d error %s", input.ID, updateResult.Error)
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			fmt.Printf("update organization RowsAffected = 0")
			return fmt.Errorf("no active organizations with id = %d", input.ID)
		}

		findResult := tx.Model(&org).Where("active = ? and id = ?", true, input.ID).Find(&org)
		if findResult.Error != nil {
			fmt.Print("after updateModel (find) error ", findResult.Error)
			return findResult.Error
		}

		tx.Exec("DELETE FROM org_teams WHERE organization_id = ?", input.ID)
		if len(input.TeamIds) > 0 {
			var teams []*Team
			tx.Model(&org).Association(
				bindModelToRelatedModels(
					tx,
					OrganizationModelName,
					TeamModelName,
					org.ID,
					input.TeamIds,
				),
			).Find(&teams)
			org.Teams = teams
		}

		tx.Exec("DELETE FROM org_users WHERE organization_id = ?", input.ID)
		if len(input.UserIds) > 0 {
			var users []*User
			tx.Model(&org).Omit("password").Association(
				bindModelToRelatedModels(
					tx,
					OrganizationModelName,
					UserModelName,
					org.ID,
					input.UserIds,
				),
			).Find(&users)
			org.Users = users
		}
		return nil
	})

	if updateError != nil {
		return nil, updateError
	}

	return org, updateError
}

func (o *Organization) GetName() string {
	return OrganizationModelName
}

func (o *Organization) ParseQueryString(endpoint string, query *gorm.DB, r *http.Request) {
	parseNameQueryParam(query, r)
}

func (o *Organization) GetSelectFields() *[]string {
	return nil
}

func (o *Organization) FindAll(query *gorm.DB) (interface{}, error) {
	records, err := findAll[Organization](nil, query)
	return records, err
}

func (o *Organization) ReadById(db *gorm.DB, id uint) (interface{}, error) {
	return readById(db, id, o)
}

func (Organization) GetAliases() []string {
	return []string{}
}

func (u *Organization) OnCreateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (u *Organization) OnReadParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (u *Organization) OnUpdateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}
