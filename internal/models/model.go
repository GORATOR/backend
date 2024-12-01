package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type GenericModel interface {
	User | Organization | Team | Project
}

type ChangableModel interface {
	CreateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error)
	UpdateModel(data []byte, userId uint, tx *gorm.DB) (interface{}, error)
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

func updateModel[T GenericModel](data []byte, tx *gorm.DB) (interface{}, error) {
	var gm T
	err := json.Unmarshal(data, &gm)
	if err != nil {
		return nil, err
	}
	insertResult := tx.Save(&gm)
	if insertResult.Error != nil {
		fmt.Print("updateModel error ", insertResult.Error)
	}
	return gm, insertResult.Error
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

func bindModelToRelatedModels(
	tx *gorm.DB,
	modelName string,
	relatedModelName string,
	modelId uint,
	relatedIDs []uint,
) string {

	for _, relatedModelId := range relatedIDs {
		query := fmt.Sprintf(
			"INSERT INTO %s_%ss (%s_id, %s_id) VALUES (?, ?)",
			getModelNameAlias(modelName),
			relatedModelName,
			modelName,
			relatedModelName,
		)
		teamBindResult := tx.Exec(query, modelId, relatedModelId)
		if teamBindResult.Error != nil {
			fmt.Printf("Bind new %s to %s with id %d error: %s", modelName, relatedModelName, modelId, teamBindResult.Error)
		}
	}
	if relatedModelName == "org" {
		return OrganizationModelName
	}
	return strings.Title(relatedModelName) + "s"
}

func bindRelatedModelsToModel(
	tx *gorm.DB,
	modelName string,
	relatedModelName string,
	modelId uint,
	relatedIDs []uint,
) string {

	for _, relatedModelId := range relatedIDs {
		query := fmt.Sprintf(
			"INSERT INTO %s_%ss (%s_id, %s_id) VALUES (?, ?)",
			getModelNameAlias(relatedModelName),
			modelName,
			relatedModelName,
			modelName,
		)
		teamBindResult := tx.Exec(query, relatedModelId, modelId)
		if teamBindResult.Error != nil {
			fmt.Printf("Bind new %s to %s with id %d error: %s", modelName, relatedModelName, modelId, teamBindResult.Error)
		}
	}
	if relatedModelName == "org" {
		return OrganizationModelName
	}
	return strings.Title(relatedModelName) + "s"
}

func getModelNameAlias(relatedModelName string) string {
	if relatedModelName == OrganizationModelName {
		return "org"
	}
	return relatedModelName
}
