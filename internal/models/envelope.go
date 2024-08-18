package models

type EnvelopeResponse struct {
	Id string `json:"id"`
}

type EventCommonSdk struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type EnvelopeRequestEventCommon struct {
	EventId string         `json:"event_id"`
	SentAt  string         `json:"sent_at"`
	DSN     string         `json:"dsn"`
	SDK     EventCommonSdk `json:"sdk"`
}

type EnvelopeRequestType struct {
	Type   string `json:"type"`
	Length string `json:"length"`
}
