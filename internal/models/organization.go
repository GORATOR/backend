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
			for _, teamId := range input.TeamIds {
				teamBindResult := tx.Exec("INSERT INTO org_teams (organization_id, team_id) VALUES (?, ?)", org.ID, teamId)
				if teamBindResult.Error != nil {
					fmt.Printf("Bind new team to organization with id %d error: %s", org.ID, teamBindResult.Error)
				}
			}

			var teams []*Team
			tx.Model(&org).Association("Teams").Find(&teams)
			org.Teams = teams
		}
		if len(input.UserIds) > 0 {
			for _, userId := range input.UserIds {
				userBindResult := tx.Exec("INSERT INTO org_users (organization_id, user_id) VALUES (?, ?)", org.ID, userId)
				if userBindResult.Error != nil {
					fmt.Printf("Bind new user to organization with id %d error: %s", org.ID, userBindResult.Error)
				}
			}

			var users []*User
			tx.Model(&org).Omit("password").Association("Users").Find(&users)
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

		findResult := tx.Model(&org).Where("active = ? and id = ?", true, input.ID).Find(&org)
		if findResult.Error != nil {
			fmt.Print("after updateModel (find) error ", findResult.Error)
			return findResult.Error
		}

		if len(input.TeamIds) > 0 {
			tx.Exec("DELETE FROM org_teams WHERE organization_id = ?", input.ID)
			for _, teamId := range input.TeamIds {
				teamBindResult := tx.Exec("INSERT INTO org_teams (organization_id, team_id) VALUES (?, ?)", input.ID, teamId)
				if teamBindResult.Error != nil {
					fmt.Printf("Bind new team to organization with id %d error: %s", input.ID, teamBindResult.Error)
				}
			}

			var teams []*Team
			tx.Model(&org).Omit("password").Association("Teams").Find(&teams)
			org.Teams = teams
		}

		if len(input.UserIds) > 0 {
			tx.Exec("DELETE FROM org_users WHERE organization_id = ?", input.ID)
			for _, userId := range input.UserIds {
				userBindResult := tx.Exec("INSERT INTO org_users (organization_id, user_id) VALUES (?, ?)", input.ID, userId)
				if userBindResult.Error != nil {
					fmt.Printf("Bind new user to organization with id %d error: %s", input.ID, userBindResult.Error)
				}
			}

			var users []*User
			tx.Model(&org).Association("Users").Find(&users)
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

func (u *Organization) OnCreateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (u *Organization) OnReadParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (u *Organization) OnUpdateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}
