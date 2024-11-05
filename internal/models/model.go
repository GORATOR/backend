package models

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type GenericModel interface {
	User | Organization | Team | Project
}

type ChangableModel interface {
	CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error)
	//UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error)
}

type Model interface {
	QueryStringParser
	InputParser
	CreatedByUser
	ChangableModel
	GetSelectFields() *[]string
	GetName() string
	FindAll(query *gorm.DB) (interface{}, error)
}

type ModelCountResponse struct {
	Entity string `json:"entity"`
	Count  int64  `json:"count"`
}

func createModel[T GenericModel](data []byte, tx *gorm.DB) (interface{}, error) {
	var gm T
	err := json.Unmarshal(data, &gm)
	if err != nil {
		return nil, err
	}
	insertResult := tx.Create(&gm)
	if insertResult.Error != nil {
		fmt.Print("createModel insert error ", insertResult.Error)
	}
	return gm, insertResult.Error
}

func findAll[T GenericModel](selectFields []string, query *gorm.DB) (interface{}, error) {
	var records []T
	if selectFields != nil {
		query.Select(selectFields)
	}
	result := query.Find(&records)
	if result.Error != nil {
		fmt.Print("tryGetRecords query.Find error ", result.Error)
		return nil, result.Error
	}
	return records, nil
}
