package models

import "encoding/json"

type ClientReportHeader struct {
	Type string `json:"type"`
}

type ClientReportDiscardedEvent struct {
	EnvelopeModel
	Reason         string `json:"reason"`
	Category       string `json:"category"`
	Quantity       uint   `json:"quantity"`
	ClientReport   ClientReport
	ClientReportID uint
}

type ClientReport struct {
	EnvelopeModel
	Project         Project
	ProjectID       uint
	EnvelopeKey     string
	Timestamp       float64                      `json:"timestamp"`
	DiscardedEvents []ClientReportDiscardedEvent `json:"discarded_events"`
}

func (c *ClientReport) GetName() string {
	return "client_report"
}

func IsClientReport(postItems []string) bool {
	if len(postItems) < 2 {
		return false
	}

	var header ClientReportHeader
	if err := json.Unmarshal([]byte(postItems[1]), &header); err != nil {
		return false
	}

	return header.Type == "client_report"
}
