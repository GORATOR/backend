package database

import (
	"errors"
	"fmt"

	"github.com/GORATOR/backend/internal/models"
	"gorm.io/gorm"
)

func EnvelopeSaveData(commonRecord *models.EnvelopeEventCommon, postItems []string, tags *[]models.EnvelopeTag) error {
	sdkResult := postgresConnection.Where(
		"name = ? and version = ?",
		commonRecord.EventCommonSdk.Name,
		commonRecord.EventCommonSdk.Version,
	).First(&commonRecord.EventCommonSdk)

	if sdkResult.Error != nil && !errors.Is(sdkResult.Error, gorm.ErrRecordNotFound) {
		return sdkResult.Error
	}
	if sdkResult.RowsAffected == 0 {
		sdkResult = postgresConnection.Create(&commonRecord.EventCommonSdk)
		if sdkResult.Error != nil {
			return sdkResult.Error
		}
	}
	commonRecord.EventCommonSdkID = &commonRecord.EventCommonSdk.ID
	commonResult := postgresConnection.Create(&commonRecord)
	if commonResult.Error != nil {
		return commonResult.Error
	}
	extraRecords := []*models.EnvelopeEventExtra{
		{
			EnvelopeEventCommonID: commonRecord.ID,
			Data:                  postItems[1],
		},
		{
			EnvelopeEventCommonID: commonRecord.ID,
			Data:                  postItems[2],
		},
	}
	extraResult := postgresConnection.Create(&extraRecords)
	if extraResult.Error != nil {
		return extraResult.Error
	}

	return bindTags(commonRecord, tags)
}

func bindTags(commonRecord *models.EnvelopeEventCommon, tags *[]models.EnvelopeTag) error {
	for _, tag := range *tags {
		findResult := postgresConnection.
			Model(&tag).
			Where("name = ? and value = ?", tag.Name, tag.Value).
			Find(&tag)
		if findResult.Error != nil {
			fmt.Println(findResult.Error)
			continue
		}
		if findResult.RowsAffected == 0 {
			tagInsertResult := postgresConnection.Create(&tag)
			if tagInsertResult.Error != nil {
				fmt.Println(findResult.Error)
				continue
			}
		}
		tag.EnvelopeEventCommon = []*models.EnvelopeEventCommon{
			commonRecord,
		}
		tagSaveResult := postgresConnection.Save(tag)
		if tagSaveResult.Error != nil {
			fmt.Printf("bindTags onSave error %s", tagSaveResult.Error)
			return tagSaveResult.Error
		}
	}

	return nil
}

func ClientReportSaveData(clientReport *models.ClientReport) error {
	result := postgresConnection.Create(clientReport)
	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("Saved client report: %d discarded events for project %d\n",
		len(clientReport.DiscardedEvents), clientReport.ProjectID)

	return nil
}
