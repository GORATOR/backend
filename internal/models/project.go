package models

import (
	"encoding/json"
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
	ProjectInput
	Active      bool
	EnvelopeKey string
	CreatedByUserStruct
}

type ProjectInput struct {
	Name   string
	TeamId uint
	Avatar string
}

func (p *Project) CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var input ProjectInput
	var insert Project
	insert.SetUserId(userId)
	err := json.Unmarshal(data, &input)
	if err != nil {
		fmt.Print("CreateModel json.Unmarshal error ", err)
		return nil, err
	}
	insert.Avatar = input.Avatar
	insert.Name = input.Name
	insert.TeamId = input.TeamId
	insert.Active = true
	insertResult := tx.Save(&insert)
	if insertResult.Error != nil {
		fmt.Print("CreateModel tx.Save error ", insertResult.Error)
	}
	return insert, insertResult.Error
}

func (p *Project) UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error) {
	var project Project
	err := json.Unmarshal(data, &project)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"name": project.Name, "avatar": project.Avatar}
	insertResult := tx.Model(&Project{}).Where("active = ? and id = ?", true, project.ID).Updates(updates)
	if insertResult.Error != nil {
		fmt.Print("updateModel error ", insertResult.Error)
	}
	tx.Model(&project).Where("active = ? and id = ?", true, project.ID).Find(&project)
	if insertResult.Error != nil {
		fmt.Print("after updateModel (find) error ", insertResult.Error)
	}
	return project, insertResult.Error
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
	parseQueryParam(query, r, "teamId", "=")
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
