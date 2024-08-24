package models

import "gorm.io/gorm"

type EnvelopeEventCommon struct {
	gorm.Model
	EventId             string
	SentAt              string
	DSN                 string
	SDK                 string
	EnvelopeEventExtras []EnvelopeEventExtra
}

type EnvelopeEventExtra struct {
	gorm.Model
	Data                  string
	EnvelopeEventCommonID uint
}
