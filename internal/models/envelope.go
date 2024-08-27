package models

import "gorm.io/gorm"

var UndefinedSdk = EventCommonSdk{
	Name:    "undefined",
	Version: "undefined",
}

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
