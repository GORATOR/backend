package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	CreatedByUserStruct
	Name   string `gorm:"index:idx_name,unique"`
	Active bool
	Users  []*User `gorm:"many2many:role_users;"`
	Rules  []*Rule
}

type RuleAction string

func (st *RuleAction) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		*st = RuleAction(v)
	case string:
		*st = RuleAction(v)
	}
	return nil
}

func (st RuleAction) Value() (driver.Value, error) {
	return string(st), nil
}

func (st RuleAction) TableName() string {
	return "public.rule_action"
}

func (st RuleAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(st))
}

func (st *RuleAction) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*st = RuleAction(s)
	return nil
}

const (
	ActionCreate RuleAction = "create"
	ActionRead   RuleAction = "read"
	ActionUpdate RuleAction = "update"
	ActionDelete RuleAction = "delete"

	RuleEntityName       = "rule"
	RuleActionEntityName = "ruleAction"
	RoleModelName        = "role"
)

type Rule struct {
	gorm.Model
	Role    Role
	RoleID  uint
	Action  RuleAction `gorm:"type:public.rule_action"`
	Allowed bool
	Table   string
}

type RuleInput struct {
	Action  RuleAction `json:"Action"`
	Allowed bool       `json:"Allowed"`
	Table   string     `json:"Table"`
}

type RoleInput struct {
	ID      uint        `json:"ID"`
	Name    string      `json:"Name"`
	UserIds []uint      `json:"UserIds"`
	Rules   []RuleInput `json:"Rules"`
}

func (r *Role) CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input RoleInput
	var role Role
	role.SetUserId(userId)
	err := json.Unmarshal(data, &input)
	if err != nil {
		fmt.Print("CreateModel json.Unmarshal error ", err)
		return nil, err
	}
	role.Active = true
	role.Name = input.Name

	insertError := tx.Transaction(func(tx *gorm.DB) error {
		insertResult := tx.Save(&role)
		if insertResult.Error != nil {
			fmt.Print("CreateModel tx.Save error ", insertResult.Error)
			return insertResult.Error
		}

		// Create rules
		if len(input.Rules) > 0 {
			for _, ruleInput := range input.Rules {
				rule := Rule{
					RoleID:  role.ID,
					Action:  ruleInput.Action,
					Allowed: ruleInput.Allowed,
					Table:   ruleInput.Table,
				}
				createResult := tx.Create(&rule)
				if createResult.Error != nil {
					fmt.Printf("CreateModel create rule error %s", createResult.Error)
					return createResult.Error
				}
			}
			// Load created rules
			tx.Where("role_id = ?", role.ID).Find(&role.Rules)
		}

		// Bind users
		if len(input.UserIds) > 0 {
			var users []*User
			tx.Model(&role).Omit("password").Association(
				bindModelToRelatedModels(
					tx,
					RoleModelName,
					UserModelName,
					role.ID,
					input.UserIds,
				),
			).Find(&users)
			role.Users = users
		}

		return nil
	})

	return role, insertError
}

func (r *Role) UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input RoleInput
	var role Role
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"name": input.Name}
	updateError := tx.Transaction(func(tx *gorm.DB) error {
		updateResult := tx.Model(&Role{}).Where("active = ? and id = ?", true, input.ID).Updates(updates)
		if updateResult.Error != nil {
			fmt.Printf("update role with id %d error %s", input.ID, updateResult.Error)
			return updateResult.Error
		}
		if updateResult.RowsAffected == 0 {
			fmt.Printf("update role RowsAffected = 0")
			return fmt.Errorf("no active roles with id = %d", input.ID)
		}

		findResult := tx.Model(&role).Where("active = ? and id = ?", true, input.ID).Find(&role)
		if findResult.Error != nil {
			fmt.Print("after updateModel (find) error ", findResult.Error)
			return findResult.Error
		}

		// Delete old rules and create new ones
		tx.Where("role_id = ?", input.ID).Delete(&Rule{})
		if len(input.Rules) > 0 {
			for _, ruleInput := range input.Rules {
				rule := Rule{
					RoleID:  role.ID,
					Action:  ruleInput.Action,
					Allowed: ruleInput.Allowed,
					Table:   ruleInput.Table,
				}
				createResult := tx.Create(&rule)
				if createResult.Error != nil {
					fmt.Printf("UpdateModel create rule error %s", createResult.Error)
					return createResult.Error
				}
			}
			// Load created rules
			tx.Where("role_id = ?", role.ID).Find(&role.Rules)
		}

		// Update users
		tx.Exec("DELETE FROM role_users WHERE role_id = ?", input.ID)
		if len(input.UserIds) > 0 {
			var users []*User
			tx.Model(&role).Omit("password").Association(
				bindModelToRelatedModels(
					tx,
					RoleModelName,
					UserModelName,
					role.ID,
					input.UserIds,
				),
			).Find(&users)
			role.Users = users
		}
		return nil
	})

	if updateError != nil {
		return nil, updateError
	}

	return role, updateError
}

func (r *Role) GetName() string {
	return RoleModelName
}

func (r *Role) ParseQueryString(endpoint string, query *gorm.DB, req *http.Request) {
	parseNameQueryParam(query, req)
}

func (r *Role) GetSelectFields() *[]string {
	return nil
}

func (r *Role) FindAll(query *gorm.DB, groupBy string) (interface{}, error) {
	records, err := findAll[Role](nil, query, groupBy)
	return records, err
}

func (r *Role) ReadById(db *gorm.DB, id uint) (interface{}, error) {
	var role Role
	result := db.Preload("Users", func(db *gorm.DB) *gorm.DB {
		return db.Omit("password")
	}).Preload("Rules").Where("id = ? and active = true", id).First(&role)
	return &role, result.Error
}

func (Role) GetAliases() []string {
	return []string{}
}

func (r *Role) OnCreateParseInput(endpoint string, query *gorm.DB, req *http.Request) error {
	return nil
}

func (r *Role) OnReadParseInput(endpoint string, query *gorm.DB, req *http.Request) error {
	return nil
}

func (r *Role) OnUpdateParseInput(endpoint string, query *gorm.DB, req *http.Request) error {
	return nil
}

func (Role) IsAllowedGroupField(groupBy string) bool {
	return isAllowedGroupFieldCommon(groupBy)
}

func (r *Role) Count(query *gorm.DB, groupBy string) (interface{}, error) {
	return countCommon(groupBy, query, r)
}
