package models

import (
	"net/http"

	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

type QueryStringParser interface {
	ParseQueryString(endpoint string, query *gorm.DB, r *http.Request)
}

type InputParser interface {
	OnCreateParseInput(query *gorm.DB, r *http.Request)
	OnReadParseInput(query *gorm.DB, r *http.Request)
	OnUpdateParseInput(query *gorm.DB, r *http.Request)
}

func parseNameQueryParam(query *gorm.DB, r *http.Request) {
	name := utils.GetQueryParam(r, "name")
	if name != "" {
		query.Where("name like ?", name+"%")
	}
}

func parseUsersQuery(query *gorm.DB, r *http.Request) {
	username := utils.GetQueryParam(r, "username")
	if username != "" {
		query.Where("username like ?", username+"%")
	}
}
