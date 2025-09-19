package models

import (
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

type QueryStringParser interface {
	ParseQueryString(endpoint string, query *gorm.DB, r *http.Request)
	IsAllowedGroupField(groupBy string) bool
}

type InputParser interface {
	OnCreateParseInput(endpoint string, query *gorm.DB, r *http.Request) error
	OnReadParseInput(endpoint string, query *gorm.DB, r *http.Request) error
	OnUpdateParseInput(endpoint string, query *gorm.DB, r *http.Request) error
}

type inSelectFunc func(id uint, db *gorm.DB) ([]uint, error)

func parseNameQueryParam(query *gorm.DB, r *http.Request) {
	param := "name"
	parseQueryParam(query, r, param, param, "like")
}

func parseUsersQuery(query *gorm.DB, r *http.Request) {
	param := "username"
	parseQueryParam(query, r, param, param, "like")
}

func parseQueryParam(
	query *gorm.DB,
	r *http.Request,
	urlParamName string,
	dbParamName string,
	sign string,
) {
	param := utils.GetQueryParam(r, urlParamName)
	if param != "" {
		query.Where(fmt.Sprintf("%s %s ?", dbParamName, sign), param+"%")
	}
}

func parseQueryParamIn(
	query *gorm.DB,
	r *http.Request,
	urlParamName string,
	inQuerySql string,
	selectFunc inSelectFunc,
) {
	id := utils.GetQueryParam(r, urlParamName)
	if id == "" {
		return
	}
	userIdUint, err := utils.StrToUint(id)
	if err != nil {
		return
	}

	ids, _ := selectFunc(userIdUint, query)
	if len(ids) > 0 {
		query.Where(inQuerySql, ids)
	}
}

func parseGroupBy(
	query *gorm.DB,
	r *http.Request,
	model QueryStringParser,
	passIdToGroup bool,
) {

	groupBy := utils.GetQueryParam(r, "groupBy")
	if groupBy == "" || !model.IsAllowedGroupField(groupBy) {
		return
	}

	if passIdToGroup {
		query.Group(fmt.Sprintf("%s, id", groupBy))
	} else {
		query.Group(groupBy)
	}

}
