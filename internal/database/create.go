package database

import (
	"errors"

	"github.com/GORATOR/backend/internal/models"
	"gorm.io/gorm"
)

func EnvelopeSaveData(commonRecord *models.EnvelopeEventCommon, postItems []string) error {
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
	commonRecord.EventCommonSdkID = commonRecord.EventCommonSdk.ID
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
	return nil
}
