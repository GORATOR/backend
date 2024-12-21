package models

import (
	"fmt"
	"net/http"
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
	ID        uint           `gorm:"primarykey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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
	EventId             string         `json:"event_id"`
	SentAt              string         `json:"sent_at"`
	DSN                 string         `json:"dsn"`
	EventCommonSdk      EventCommonSdk `json:"sdk"`
	EventCommonSdkID    uint           `json:"-"`
	EnvelopeEventExtras []EnvelopeEventExtra
	EnvelopeKey         string `json:"-"`
	Project             Project
	ProjectID           uint
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
	e.EnvelopeKey = string([]rune(e.DSN)[first+1 : second])
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

func (e *EnvelopeEventCommon) FindAll(query *gorm.DB) (interface{}, error) {
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

func (e *EnvelopeEventCommon) GetName() string {
	return EnvelopeEventCommonModelName
}

func (e *EnvelopeEventCommon) GetSelectFields() *[]string {
	return nil
}

func (e *EnvelopeEventCommon) ParseQueryString(endpoint string, query *gorm.DB, r *http.Request) {
	parseQueryParam(query, r, "project_id", "project_id", "=")
	parseQueryParam(query, r, "created_at_from", "created_at", ">=")
	parseQueryParam(query, r, "created_at_to", "created_at", "<=")
	userId := utils.GetQueryParam(r, "user_id")
	if userId != "" {
		//todo: п3 Доступные пользователю (пользователь связан с проектом, а конверты с проектом)
		// найти все проекты, которые связаны с командами, в которые добавлен user_id, использовать в in их project_id
	}
	teamId := utils.GetQueryParam(r, "team_id")
	if teamId != "" {
		//todo: п4 Доступные команде (проект связан с командой , а конверты с проектом)
		// найти все проекты, которые связаны с командой, использовать в in их project_id
	}
}
