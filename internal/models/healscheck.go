package models

type Healthcheck struct {
	Version         string `json:"version"`
	Status          string `json:"status"`
	PostgresVersion string `json:"postgresVersion"`
}
