package models

import "encoding/json"

const TransactionModelName = "transaction"

type Transaction struct {
	EnvelopeModel
	EventId        string   `json:"event_id" gorm:"type:text"`
	TraceId        string   `json:"trace_id" gorm:"type:text"`
	SpanId         string   `json:"span_id" gorm:"type:text"`
	ParentSpanId   string   `json:"parent_span_id" gorm:"type:text"`
	Name           string   `json:"name" gorm:"type:text"`
	Op             string   `json:"op" gorm:"type:text"`
	Status         string   `json:"status" gorm:"type:text"`
	StartTimestamp string   `json:"start_timestamp" gorm:"column:start_timestamp;type:text"`
	EndTimestamp   string   `json:"end_timestamp" gorm:"column:end_timestamp;type:text"`
	ProjectID      *uint    `json:"project_id"`
	Project        *Project `json:"project,omitempty"`
	Spans          []Span   `json:"spans" gorm:"foreignKey:TransactionID"`
	Contexts       *string  `json:"contexts" gorm:"type:jsonb"`
	Tags           *string  `json:"tags" gorm:"type:jsonb"`
}

type Span struct {
	EnvelopeModel
	SpanId         string  `json:"span_id" gorm:"type:text"`
	ParentSpanId   string  `json:"parent_span_id" gorm:"type:text"`
	TraceId        string  `json:"trace_id" gorm:"type:text"`
	TransactionID  uint    `json:"transaction_id"`
	Op             string  `json:"op" gorm:"type:text"`
	Description    string  `json:"description" gorm:"type:text"`
	Status         string  `json:"status" gorm:"type:text"`
	StartTimestamp string  `json:"start_timestamp" gorm:"column:start_timestamp;type:text"`
	EndTimestamp   string  `json:"end_timestamp" gorm:"column:end_timestamp;type:text"`
	Data           *string `json:"data" gorm:"type:jsonb"`
}

// FindEnvelopeItemPayload returns the payload string for the first envelope item
// matching itemType. Envelope items are (header, payload) pairs at positions 1,2 / 3,4 / ...
func FindEnvelopeItemPayload(postItems []string, itemType string) (string, bool) {
	for i := 1; i+1 < len(postItems); i += 2 {
		var h struct {
			Type string `json:"type"`
		}
		if json.Unmarshal([]byte(postItems[i]), &h) == nil && h.Type == itemType {
			return postItems[i+1], true
		}
	}
	return "", false
}

func IsTransaction(postItems []string) bool {
	_, ok := FindEnvelopeItemPayload(postItems, "transaction")
	return ok
}
