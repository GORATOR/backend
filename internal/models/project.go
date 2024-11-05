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
	Active      bool
	Avatar      string
	EnvelopeKey string
	CreatedByUserStruct
}

func (p *Project) CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	return createModel[Project](data, tx)
}

func (p *Project) GenerateEnvelopeKey() {
	if p.ID > 0 {
		return
	}
	now := time.Now().Unix()
	data := fmt.Sprintf(
		"%s %d %s",
		p.Name,
		now,
		utils.RandomString(7),
	)
	p.EnvelopeKey = utils.GenerateMd5(data)
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

func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	p.GenerateEnvelopeKey()
	return nil
}

func (p *Project) OnCreateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (p *Project) OnReadParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}

func (p *Project) OnUpdateParseInput(endpoint string, query *gorm.DB, r *http.Request) error {
	return nil
}
