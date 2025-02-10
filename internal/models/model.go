package models

import (
	"encoding/json"
	"fmt"
	"slices"
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
	ReadableModel
	WritebleModel
}

type Countable interface {
	Count(query *gorm.DB, groupBy string) (interface{}, error)
}

type ReadableModel interface {
	ModelCommon
	QueryStringParser
	Countable
	FindAll(query *gorm.DB, groupBy string) (interface{}, error)
	ReadById(db *gorm.DB, id uint) (interface{}, error)
}

type WritebleModel interface {
	ModelCommon
	CreatedByUser
	ChangableModel
}

type ModelCommon interface {
	GetName() string
	GetAliases() []string
	GetSelectFields() *[]string
}

type ModelCountResponse struct {
	Entity string `json:"entity"`
	Count  int64  `json:"count"`
}

type ModelGroupedCountRecord struct {
	Field string `json:"field"`
	Count int64  `json:"count"`
}

type ModelGroupedCountResponse struct {
	ModelCountResponse
	GroupBy     string                    `json:"groupBy"`
	GroupedData []ModelGroupedCountRecord `json:"groupedData"`
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

func findAll[T GenericModel](selectFields []string, query *gorm.DB, groupBy string) (interface{}, error) {
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

func readById(db *gorm.DB, id uint, m ReadableModel) (interface{}, error) {
	result := db.Where("id = ? and active = true", id).First(&m)
	return &m, result.Error
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

func isAllowedGroupFieldCommon(groupBy string) bool {
	return slices.Contains(
		[]string{
			"name",
			"created_at",
			"updated_at",
			"created_by",
		},
		groupBy,
	)
}

func countCommon(groupBy string, query *gorm.DB, m ReadableModel) (interface{}, error) {
	if groupBy != "" {
		return countEntitiesGroupedResult(groupBy, query, m)
	} else {
		return countEntitiesResult(query, m)
	}
}

func countEntitiesGroupedResult(groupBy string, query *gorm.DB, m ReadableModel) (ModelGroupedCountResponse, error) {
	response := ModelGroupedCountResponse{
		ModelCountResponse: ModelCountResponse{
			Entity: m.GetName(),
			Count:  0,
		},
		GroupBy:     groupBy,
		GroupedData: []ModelGroupedCountRecord{},
	}
	var result []ModelGroupedCountRecord
	var count int64
	if !m.IsAllowedGroupField(groupBy) {
		fmt.Printf("countEntitiesGroupedResult: using disallowed field %s", groupBy)
		return response, fmt.Errorf("using disallowed field %s", groupBy)
	}
	selectStr := fmt.Sprintf("count(*) AS count, %s AS field", groupBy)
	countResult := query.Select(selectStr).Scan(&result)

	if countResult.Error != nil {
		return response, countResult.Error
	}

	count = 0
	for _, rec := range result {
		count += rec.Count
	}
	response.GroupedData = result
	response.ModelCountResponse.Count = count

	return response, nil

}

func countEntitiesResult(query *gorm.DB, m ReadableModel) (ModelCountResponse, error) {
	response := ModelCountResponse{
		Entity: m.GetName(),
		Count:  0,
	}
	var count int64
	countResult := query.Count(&count)

	if countResult.Error != nil {
		return response, countResult.Error
	}

	response.Count = count
	return response, nil
}
