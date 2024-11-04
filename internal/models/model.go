package models

import (
	"fmt"

	"gorm.io/gorm"
)

type GenericModel interface {
	User | Organization | Team | Project
}

type Model interface {
	InputParser
	GetSelectFields() *[]string
	GetName() string
	FindAll(query *gorm.DB) (interface{}, error)
}

type ModelCountResponse struct {
	Entity string `json:"entity"`
	Count  int64  `json:"count"`
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
