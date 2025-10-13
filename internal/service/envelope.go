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
				if commonRecord.EventCommonSdk == nil {
					commonRecord.EventCommonSdk = &models.EventCommonSdk{}
				}
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
		commonRecord.EventCommonSdk = &models.EventCommonSdk{
			Name:    models.UndefinedSdk.Name,
			Version: models.UndefinedSdk.Version,
		}
	}
	return nil
}

func ParseTags(postItems []string) ([]models.EnvelopeTag, error) {
	var p fastjson.Parser
	result := []models.EnvelopeTag{}
	v, err := p.Parse(postItems[models.EnvelopePostItemMessage])
	if err != nil {
		fmt.Println(err)
	} else {
		if v.Exists("tags") {
			t := v.GetObject("tags")
			t.Visit(tagVisit(&result))
			return result, nil
		}
	}
	return result, nil
}

func isSdkEmpty(commonRecord *models.EnvelopeEventCommon) bool {
	return commonRecord.EventCommonSdk == nil || (commonRecord.EventCommonSdk.Version == "" && commonRecord.EventCommonSdk.Name == "")
}

func formatValue(val string) string {
	return strings.ReplaceAll(val, "\"", "")
}

func tagVisit(tags *[]models.EnvelopeTag) func(key []byte, v *fastjson.Value) {
	return func(key []byte, v *fastjson.Value) {
		*tags = append(
			*tags,
			models.EnvelopeTag{
				Name:  string(key),
				Value: strings.Trim(v.String(), "\""),
			},
		)
	}
}

func ParseException(commonRecord *models.EnvelopeEventCommon, postItems []string) error {
	if len(postItems) < 3 {
		return nil
	}

	var p fastjson.Parser
	v, err := p.Parse(postItems[models.EnvelopePostItemMessage])
	if err != nil {
		return nil
	}

	if !v.Exists("exception") {
		return nil
	}

	exceptionObj := v.Get("exception")
	if exceptionObj == nil {
		return nil
	}

	// Store full exception data as JSONB
	commonRecord.ExceptionData = exceptionObj.String()

	// Extract type and value for indexing
	var exceptionValue *fastjson.Value

	// Try format: exception[0]
	if exceptionObj.Type() == fastjson.TypeArray {
		arr := exceptionObj.GetArray()
		if len(arr) > 0 {
			exceptionValue = arr[0]
		}
	} else if exceptionObj.Type() == fastjson.TypeObject {
		// Try format: exception.values[0]
		values := exceptionObj.Get("values")
		if values != nil && values.Type() == fastjson.TypeArray {
			arr := values.GetArray()
			if len(arr) > 0 {
				exceptionValue = arr[0]
			}
		}
	}

	if exceptionValue != nil {
		typeVal := exceptionValue.Get("type")
		valueVal := exceptionValue.Get("value")

		if typeVal != nil {
			commonRecord.ExceptionType = formatValue(typeVal.String())
		}
		if valueVal != nil {
			commonRecord.ExceptionValue = formatValue(valueVal.String())
		}
	}

	return nil
}

func ParseClientReport(postItems []string) (*models.ClientReport, error) {
	if len(postItems) < 3 {
		return nil, fmt.Errorf("invalid client report format")
	}

	var report models.ClientReport
	if err := json.Unmarshal([]byte(postItems[2]), &report); err != nil {
		return nil, fmt.Errorf("failed to parse client report data: %v", err)
	}

	return &report, nil
}
