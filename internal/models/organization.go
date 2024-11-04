package models

import (
	"net/http"

	"gorm.io/gorm"
)

const (
	OrganizationModelName = "organization"
)

type Organization struct {
	gorm.Model
	Name   string
	Avatar string
	Active bool
	Teams  []*Team `gorm:"many2many:org_teams;"`
	Users  []*User `gorm:"many2many:org_users;"`
}

func (o *Organization) GetName() string {
	return OrganizationModelName
}

func (o *Organization) ParseInput(query *gorm.DB, r *http.Request) {
	parseNameQueryParam(query, r)
}

func (o *Organization) GetSelectFields() *[]string {
	return nil
}

func (o *Organization) FindAll(query *gorm.DB) (interface{}, error) {
	records, err := findAll[Organization](nil, query)
	return records, err
}
