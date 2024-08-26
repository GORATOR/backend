package database

import (
	"github.com/GORATOR/backend/internal/models"
)

func EnvelopeSaveData(commonRecord *models.EnvelopeEventCommon, postItems []string) error {
	sdkResult := postgresConnection.Where(
		&commonRecord.EventCommonSdk,
		"name = ? and version = ?",
		commonRecord.EventCommonSdk.Model,
		commonRecord.EventCommonSdk.Version,
	)
	if sdkResult.Error != nil {
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
