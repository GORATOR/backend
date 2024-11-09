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
	CreatedByUserStruct
	Name   string
	Avatar string
	Active bool
	Teams  []*Team `gorm:"many2many:org_teams;"`
	Users  []*User `gorm:"many2many:org_users;"`
}

func (o *Organization) CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	return createModel[Organization](data, tx)
}

func (o *Organization) UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	return updateModel[Organization](data, tx)
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
