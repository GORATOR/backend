package models

import (
	"fmt"
	"net/http"
	"time"

	"github.com/GORATOR/backend/internal/utils"
	"gorm.io/gorm"
)

const (
	ProjectModelName = "project"
)

type Project struct {
	gorm.Model
	Name        string
	TeamId      uint
	Team        Team `gorm:"foreignKey:TeamId"`
	Active      bool
	Avatar      string
	EnvelopeKey string
	CreatedBy   uint
	User        User `gorm:"foreignKey:CreatedBy"`
}

func (p *Project) GenerateEnvelopeKey() {
	if p.ID > 0 {
		now := time.Now().Unix()
		data := fmt.Sprintf(
			"%s %s %s %d %s",
			p.Name, p.Team.Name,
			p.User.Username,
			now,
			utils.RandomString(7),
		)
		p.EnvelopeKey = utils.GenerateMd5(data)
	}
}

func (p *Project) GetName() string {
	return ProjectModelName
}

func (p *Project) ParseQueryString(endpoint string, query *gorm.DB, r *http.Request) {
	parseNameQueryParam(query, r)
}

func (p *Project) GetSelectFields() *[]string {
	return nil
}

func (p *Project) FindAll(query *gorm.DB) (interface{}, error) {
	records, err := findAll[Project](nil, query)
	return records, err
}
