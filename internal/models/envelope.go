package models

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

var UndefinedSdk = EventCommonSdk{
	Name:    "undefined",
	Version: "undefined",
}

const (
	EnvelopePostItemCommon  = 0
	EnvelopePostItemType    = 1
	EnvelopePostItemMessage = 2

	EnvelopeRequiredPostItems    = 3
	EnvelopeKeyFormatError       = "wrong DSN format (%s instead of something like http://KEY@DOMAIN:PORT/2)"
	EnvelopeKeyWrongHeaderError  = "wrong X-Sentry-Auth header format"
	EnvelopeEventCommonModelName = "envelope_event_common"
)

type EnvelopeModel struct {
	ID        uint `gorm:"primarykey" `
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" `
}

type EnvelopeTag struct {
	EnvelopeModel
	EnvelopeEventCommon []*EnvelopeEventCommon `gorm:"many2many:eetags_eecommon;"`
	Name                string
	Value               string
}

type EnvelopeResponse struct {
	Id string `json:"id"`
}

type EventCommonSdk struct {
	EnvelopeModel
	Name    string `json:"name"`
	Version string `json:"version"`
}

type EnvelopeEventCommon struct {
	EnvelopeModel
	EventId             string               `json:"event_id"`
	SentAt              string               `json:"sent_at"`
	DSN                 string               `json:"dsn"`
	EventCommonSdk      *EventCommonSdk      `json:"sdk"`
	EventCommonSdkID    *uint                `json:"-"`
	EnvelopeEventExtras []EnvelopeEventExtra `json:"EnvelopeEventExtras"`
	EnvelopeKey         string               `json:"-"`
	Project             *Project             `json:"project"`
	ProjectID           *uint                `json:"ProjectID"`
}

type EnvelopeRequestType struct {
	Type   string `json:"type"`
	Length string `json:"length"`
}

type EnvelopeEventExtra struct {
	EnvelopeModel
	Data                  string
	EnvelopeEventCommonID uint
}

func (e *EnvelopeEventCommon) TryExtractKeyFromDsn() error {
	if len(e.DSN) == 0 {
		return fmt.Errorf("empty DSN")
	}
	first := strings.Index(e.DSN, "//")
	if first <= 0 {
		return fmt.Errorf(EnvelopeKeyFormatError, e.DSN)
	}
	second := strings.Index(e.DSN, "@")
	if second <= 0 || first > second {
		return fmt.Errorf(EnvelopeKeyFormatError, e.DSN)
	}
	e.EnvelopeKey = string([]rune(e.DSN)[first+2 : second])
	return nil
}

func (e *EnvelopeEventCommon) TryExtractKeyFromHeaders(r *http.Request) error {
	sentryHeader := r.Header.Get("X-Sentry-Auth")
	if len(sentryHeader) == 0 {
		return fmt.Errorf("empty DSN")
	}
	sentryParams := strings.Split(strings.ReplaceAll(sentryHeader, "Sentry", ""), ",")
	if len(sentryParams) == 0 {
		return fmt.Errorf(EnvelopeKeyWrongHeaderError)
	}
	for _, sentryParam := range sentryParams {
		separatedData := strings.Split(strings.TrimSpace(sentryParam), "=")
		if len(separatedData) != 2 || separatedData[0] != "sentry_key" {
			continue
		}
		e.EnvelopeKey = separatedData[1]
		return nil
	}
	return fmt.Errorf(EnvelopeKeyWrongHeaderError)
}

func (e *EnvelopeEventCommon) TryExtractKeyFromUri(r *http.Request) error {
	queryParams := r.URL.Query()
	if !queryParams.Has("sentry_key") {
		return fmt.Errorf("empty DSN")
	}
	e.EnvelopeKey = queryParams.Get("sentry_key")
	return nil
}

func (e *EnvelopeEventCommon) TryExtractKey(r *http.Request) error {
	err := e.TryExtractKeyFromDsn()
	if err == nil {
		return nil
	}
	err = e.TryExtractKeyFromHeaders(r)
	if err == nil {
		return nil
	}
	err = e.TryExtractKeyFromUri(r)

	return err
}

func (e *EnvelopeEventCommon) FindAll(query *gorm.DB, groupBy string) (interface{}, error) {
	if groupBy == "tag" {
		return e.findAllGroupByTag(query)
	}
	var records []EnvelopeEventCommon
	result := query.Find(&records)
	if result.Error != nil {
		fmt.Print("tryGetRecords query.Find error ", result.Error)
		return nil, result.Error
	}
	if len(records) == 0 {
		return records, nil
	}
	query.
		Preload("EventCommonSdk").
		Preload("EnvelopeEventExtras").
		Preload("Project").
		Find(&records)
	return records, nil
}

func (EnvelopeEventCommon) findAllGroupByTag(query *gorm.DB) (interface{}, error) {
	var scanResult []struct {
		EnvelopeEventCommon
		Tag string
	}
	result := make(map[string][]EnvelopeEventCommon)
	findResult := query.Scan(&scanResult)
	if findResult.Error != nil {
		return nil, findResult.Error
	}
	for _, row := range scanResult {
		result[row.Tag] = append(result[row.Tag], row.EnvelopeEventCommon)
	}
	return result, nil
}

func (e *EnvelopeEventCommon) ReadById(db *gorm.DB, id uint) (interface{}, error) {
	var record EnvelopeEventCommon
	result := db.Where("id = ?", id).
		Preload("EventCommonSdk").
		Preload("EnvelopeEventExtras").
		Preload("Project").
		First(&record)
	return &record, result.Error
}

func (e *EnvelopeEventCommon) GetName() string {
	return EnvelopeEventCommonModelName
}

func (e *EnvelopeEventCommon) GetSelectFields() *[]string {
	return nil
}

func (EnvelopeEventCommon) GetAliases() []string {
	return []string{
		"issue",
		"envelope",
	}
}

func (e *EnvelopeEventCommon) ParseQueryString(endpoint string, query *gorm.DB, r *http.Request) {
	parseQueryParam(query, r, "projectId", "project_id", "=")
	parseQueryParam(query, r, "createdAtFrom", "created_at", ">=")
	parseQueryParam(query, r, "createdAtTo", "created_at", "<=")

	parseQueryParamIn(query, r, "userId", "project_id IN ?", getUserProjectIDs)
	parseQueryParamIn(query, r, "teamId", "project_id IN ?", getTeamProjectIDs)

	e.parseGroupBy(query, r)
}

func (EnvelopeEventCommon) IsAllowedGroupField(groupBy string) bool {
	return slices.Contains(
		[]string{
			"name",
			"created_at",
			"updated_at",
			"sent_at",
			"dsn",
			"project_id",
			"envelope_key",
			"tag",
		},
		groupBy,
	)
}

func (e *EnvelopeEventCommon) parseGroupBy(query *gorm.DB, r *http.Request) {
	groupBy := utils.GetQueryParam(r, "groupBy")
	if groupBy == "" {
		return
	}

	if groupBy == "tag" {
		if r.URL.Path == "/envelopes/count" {
			query.Table("envelope_event_commons AS eec").
				Select("count(eec.*) as count, et.value AS field").
				Joins("LEFT JOIN eetags_eecommon ee ON ee.envelope_event_common_id = eec.id AND eec.id IS NOT NULL").
				Joins("LEFT JOIN envelope_tags et ON et.id = ee.envelope_tag_id AND et.id IS NOT NULL").
				Group("et.value")
		} else {
			query.Table("envelope_event_commons AS eec").
				Select("eec.*, et.value AS tag").
				Joins("LEFT JOIN eetags_eecommon ee ON ee.envelope_event_common_id = eec.id AND eec.id IS NOT NULL").
				Joins("LEFT JOIN envelope_tags et ON et.id = ee.envelope_tag_id AND et.id IS NOT NULL").
				Group("et.value, eec.id")
		}
	} else {
		query.Group(groupBy)
	}

}

func (e *EnvelopeEventCommon) countEntitiesGroupedByTag(groupBy string, query *gorm.DB, m ReadableModel) (ModelGroupedCountResponse, error) {
	response := ModelGroupedCountResponse{
		ModelCountResponse: ModelCountResponse{
			Entity: m.GetName(),
			Count:  0,
		},
		GroupBy:     groupBy,
		GroupedData: []ModelGroupedCountRecord{},
	}
	var result []ModelGroupedCountRecord
	var count int64
	if !m.IsAllowedGroupField(groupBy) {
		fmt.Printf("countEntitiesGroupedResult: using disallowed field %s", groupBy)
		return response, fmt.Errorf("using disallowed field %s", groupBy)
	}
	countResult := query.Scan(&result)

	if countResult.Error != nil {
		return response, countResult.Error
	}

	count = 0
	for _, rec := range result {
		count += rec.Count
	}
	response.GroupedData = result
	response.ModelCountResponse.Count = count

	return response, nil

}

func (e *EnvelopeEventCommon) Count(query *gorm.DB, groupBy string) (interface{}, error) {
	if groupBy == "" {
		return countEntitiesResult(query, e)
	}
	if groupBy == "tag" {
		return e.countEntitiesGroupedByTag(groupBy, query, e)
	}
	return countEntitiesGroupedResult(groupBy, query, e)
}
