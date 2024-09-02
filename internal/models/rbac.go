package models

import (
	"database/sql/driver"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name  string
	Users []*User `gorm:"many2many:role_users;"`
	Rules []*Rule
}

type RuleAction string

func (st *RuleAction) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		*st = RuleAction(b)
	}
	return nil
}

func (st RuleAction) Value() (driver.Value, error) {
	return string(st), nil
}

func (st RuleAction) TableName() string {
	return "public.rule_action"
}

const (
	Create RuleAction = "create"
	Read   RuleAction = "read"
	Update RuleAction = "update"
	Delete RuleAction = "delete"
)

type Rule struct {
	gorm.Model
	Role    Role
	RoleID  uint
	Action  RuleAction `gorm:"type:public.rule_action"`
	Allowed bool
	Table   string
}
