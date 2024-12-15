package models

import (
	"fmt"
	"strings"

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

	EnvelopeRequiredPostItems = 3
	EnvelopeKeyFormatError    = "wrong DSN format (%s instead of something like http://KEY@DOMAIN:PORT/2)"
)

type EnvelopeResponse struct {
	Id string `json:"id"`
}

type EventCommonSdk struct {
	gorm.Model
	Name    string `json:"name"`
	Version string `json:"version"`
}

type EnvelopeEventCommon struct {
	gorm.Model
	EventId             string         `json:"event_id"`
	SentAt              string         `json:"sent_at"`
	DSN                 string         `json:"dsn"`
	EventCommonSdk      EventCommonSdk `json:"sdk"`
	EventCommonSdkID    uint
	EnvelopeEventExtras []EnvelopeEventExtra
	EnvelopeKey         string
}

type EnvelopeRequestType struct {
	Type   string `json:"type"`
	Length string `json:"length"`
}

type EnvelopeEventExtra struct {
	gorm.Model
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
