package models

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

const (
	TeamModelName = "team"
)

type Team struct {
	gorm.Model
	CreatedByUserStruct
	Name          string
	Avatar        string
	Active        bool
	Users         []*User         `gorm:"many2many:team_users;"`
	Organizations []*Organization `gorm:"many2many:org_teams;"`
	Projects      []*Project      `gorm:"foreignKey:TeamId"`
}

type TeamInput struct {
	ID              uint
	Name            string
	Avatar          string
	UserIds         []uint
	OrganizationIds []uint
	ProjectIds      []uint
}

func (t *Team) CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input TeamInput
	var team Team
	team.SetUserId(userId)
	err := json.Unmarshal(data, &input)
	if err != nil {
		fmt.Print("CreateModel json.Unmarshal error ", err)
		return nil, err
	}
	team.Avatar = input.Avatar
	team.Name = input.Name
	team.Active = true

	insertError := tx.Transaction(func(tx *gorm.DB) error {
		insertResult := tx.Save(&team)
		if insertResult.Error != nil {
			fmt.Print("CreateModel tx.Save error ", insertResult.Error)
			return insertResult.Error
		}

		if len(input.UserIds) > 0 {
			var users []*User
			tx.Model(&team).Omit("password").Association(
				bindModelToRelatedModels(
					tx,
					TeamModelName,
					UserModelName,
					team.ID,
					input.UserIds,
				),
			).Find(&users)
			team.Users = users
		}

		if len(input.OrganizationIds) > 0 {
			var orgs []*Organization
			tx.Model(&team).Association(
				bindRelatedModelsToModel(
					tx,
					TeamModelName,
					OrganizationModelName,
					team.ID,
					input.OrganizationIds,
				),
			).Find(&orgs)
			team.Organizations = orgs
		}

		if len(input.ProjectIds) > 0 {
			updates := map[string]interface{}{"team_id": team.ID}
			tx.Model(&Project{}).
				Where("id IN ?", input.ProjectIds).
				Updates(updates)
			var projects []*Project
			tx.Model(&Project{}).Where("team_id = ?", team.ID).Find(&projects)
			team.Projects = projects
		}

		return nil
	})

	return team, insertError
}

func (t *Team) UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input TeamInput
	var team Team
	team.SetUserId(userId)
	err := json.Unmarshal(data, &input)
	if err != nil {
		fmt.Print("CreateModel json.Unmarshal error ", err)
		return nil, err
	}
	updates := map[string]interface{}{
		"name":   input.Name,
		"avatar": input.Avatar,
	}

	updateError := tx.Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&Team{}).Where("active = ? and id = ?", true, input.ID).Updates(updates)
		if updateResult.Error != nil {
			fmt.Printf("update team with id %d error %s", input.ID, updateResult.Error)
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			fmt.Printf("update team RowsAffected = 0")
			return fmt.Errorf("no active teams with id = %d", input.ID)
		}

		findResult := tx.Model(&team).Where("active = ? and id = ?", true, input.ID).Find(&team)
		if findResult.Error != nil {
			fmt.Print("after updateModel (find) error ", findResult.Error)
			return findResult.Error
		}

		tx.Exec("DELETE FROM org_teams WHERE team_id = ?", input.ID)
		if len(input.OrganizationIds) > 0 {
			var orgs []*Organization
			tx.Model(&team).Association(
				bindRelatedModelsToModel(
					tx,
					TeamModelName,
					OrganizationModelName,
					team.ID,
					input.OrganizationIds,
				),
			).Find(&orgs)
			team.Organizations = orgs
		}

		tx.Exec("DELETE FROM team_users WHERE team_id = ?", input.ID)
		if len(input.UserIds) > 0 {
			var users []*User
			tx.Model(&team).Omit("password").Association(
				bindModelToRelatedModels(
					tx,
					TeamModelName,
					UserModelName,
					team.ID,
					input.UserIds,
				),
			).Find(&users)
			team.Users = users
		}

		tx.Exec("UPDATE projects SET team_id = null WHERE team_id = ?", input.ID)
		if len(input.ProjectIds) > 0 {
			updates := map[string]interface{}{"team_id": input.ID}
			tx.Model(&Project{}).
				Where("id IN ?", input.ProjectIds).
				Updates(updates)
			var projects []*Project
			tx.Model(&Project{}).Where("team_id = ?", input.ID).Find(&projects)
			team.Projects = projects
		}

		return nil
	})

	if updateError != nil {
		return nil, updateError
	}
	return team, updateError
}

func (t *Team) GetName() string {
	return TeamModelName
}

func (t *Team) ParseQueryString(endpoint string, query *gorm.DB, r *http.Request) {
	parseNameQueryParam(query, r)
	parseGroupBy(query, r)
}

func (t *Team) GetSelectFields() *[]string {
	return nil
}

func (t *Team) FindAll(query *gorm.DB) (interface{}, error) {
	records, err := findAll[Team](nil, query)
	return records, err
}

func (t *Team) ReadById(db *gorm.DB, id uint) (interface{}, error) {
	return readById(db, id, t)
}

func (Team) GetAliases() []string {
	return []string{}
}

func (u *Team) OnCreateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (u *Team) OnReadParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (u *Team) OnUpdateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (Team) IsAllowedGroupField(groupBy string) bool {
	return isAllowedGroupFieldCommon(groupBy)
}
