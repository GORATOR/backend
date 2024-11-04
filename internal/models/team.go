package models

import (
	"net/http"

	"gorm.io/gorm"
)

const (
	TeamModelName = "team"
)

type Team struct {
	gorm.Model
	Name          string
	Avatar        string
	Active        bool
	Users         []*User         `gorm:"many2many:team_users;"`
	Organizations []*Organization `gorm:"many2many:org_teams;"`
}

func (t *Team) GetName() string {
	return TeamModelName
}

func (t *Team) ParseQueryString(endpoint string, query *gorm.DB, r *http.Request) {
	parseNameQueryParam(query, r)
}

func (t *Team) GetSelectFields() *[]string {
	return nil
}

func (t *Team) FindAll(query *gorm.DB) (interface{}, error) {
	records, err := findAll[Team](nil, query)
	return records, err
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
