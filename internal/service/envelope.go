package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GORATOR/backend/internal/models"
	"github.com/valyala/fastjson"
)

func ParseSDK(commonRecord *models.EnvelopeEventCommon, postItems []string) error {
	if err := json.Unmarshal([]byte(postItems[models.EnvelopePostItemCommon]), &commonRecord); err != nil {
		return err
	}
	if isSdkEmpty(commonRecord) {
		var p fastjson.Parser
		v, err := p.Parse(postItems[models.EnvelopePostItemMessage])
		if err != nil {
			fmt.Println(err)
		} else {
			if v.Exists("sdk") {
				sdkObject := v.GetObject("sdk")
				name := sdkObject.Get("name")
				version := sdkObject.Get("version")
				if name != nil {
					commonRecord.EventCommonSdk.Name = formatValue(name.String())
				}
				if version != nil {
					commonRecord.EventCommonSdk.Version = formatValue(version.String())
				}
			}
		}
	}

	if isSdkEmpty(commonRecord) {
		commonRecord.EventCommonSdk = models.EventCommonSdk{
			Name:    models.UndefinedSdk.Name,
			Version: models.UndefinedSdk.Version,
		}
	}
	return nil
}

func isSdkEmpty(commonRecord *models.EnvelopeEventCommon) bool {
	return commonRecord.EventCommonSdk.Version == "" && commonRecord.EventCommonSdk.Name == ""
}

func formatValue(val string) string {
	return strings.ReplaceAll(val, "\"", "")
}
