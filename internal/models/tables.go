package models

import "gorm.io/gorm"

type EnvelopeEventCommon struct {
	gorm.Model
	Id      uint
	EventId string
	SentAt  string
	DSN     string
	SDK     string
}

type EnvelopeEventExtra struct {
	gorm.Model
	Id     uint
	Common EnvelopeEventCommon
	Data   string
}
