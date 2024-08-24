package database

import (
	"encoding/json"

	"github.com/GORATOR/backend/internal/models"
)

func EnvelopeSaveData(commonData *models.EnvelopeRequestEventCommon, postItems []string) error {
	sdkBytes, err := json.Marshal(commonData.SDK)
	if err != nil {
		return err
	}

	commonRecord := models.EnvelopeEventCommon{
		SentAt:  commonData.SentAt,
		DSN:     commonData.DSN,
		EventId: commonData.EventId,
		SDK:     string(sdkBytes),
	}
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
