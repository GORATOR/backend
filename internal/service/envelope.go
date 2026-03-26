package service

import (
	"crypto/sha256"
	"encoding/hex"
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
					commonRecord.EventCommonSdk.Name = string(name.GetStringBytes())
				}
				if version != nil {
					commonRecord.EventCommonSdk.Version = string(version.GetStringBytes())
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

func tagVisit(tags *[]models.EnvelopeTag) func(key []byte, v *fastjson.Value) {
	return func(key []byte, v *fastjson.Value) {
		*tags = append(
			*tags,
			models.EnvelopeTag{
				Name:  string(key),
				Value: string(v.GetStringBytes()),
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
	exceptionData := exceptionObj.String()
	commonRecord.ExceptionData = &exceptionData

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
			commonRecord.ExceptionType = string(typeVal.GetStringBytes())
		}
		if valueVal != nil {
			commonRecord.ExceptionValue = string(valueVal.GetStringBytes())
		}
	}

	commonRecord.StacktraceHash = ComputeStacktraceHash(exceptionObj)

	return nil
}

// ComputeStacktraceHash builds a fingerprint from exception type/module and frame
// structure (filename, function, lineno) — variable values are excluded.
func ComputeStacktraceHash(exceptionObj *fastjson.Value) string {
	if exceptionObj == nil {
		return ""
	}

	var parts []string

	var values []*fastjson.Value
	if exceptionObj.Type() == fastjson.TypeArray {
		values = exceptionObj.GetArray()
	} else if exceptionObj.Type() == fastjson.TypeObject {
		v := exceptionObj.Get("values")
		if v != nil && v.Type() == fastjson.TypeArray {
			values = v.GetArray()
		}
	}

	for _, exc := range values {
		excType := string(exc.GetStringBytes("type"))
		excModule := string(exc.GetStringBytes("module"))
		parts = append(parts, excType+":"+excModule)

		stacktrace := exc.Get("stacktrace")
		if stacktrace == nil {
			continue
		}
		frames := stacktrace.GetArray("frames")
		for _, frame := range frames {
			filename := string(frame.GetStringBytes("filename"))
			function := string(frame.GetStringBytes("function"))
			lineno := frame.GetInt("lineno")
			parts = append(parts, fmt.Sprintf("%s:%s:%d", filename, function, lineno))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func ParseExtra(commonRecord *models.EnvelopeEventCommon, postItems []string) error {
	if len(postItems) < 3 {
		return nil
	}

	var p fastjson.Parser
	v, err := p.Parse(postItems[models.EnvelopePostItemMessage])
	if err != nil {
		return nil
	}

	if !v.Exists("extra") {
		return nil
	}

	extraObj := v.Get("extra")
	if extraObj == nil {
		return nil
	}

	extraData := extraObj.String()
	commonRecord.ExtraData = &extraData
	return nil
}

func ParseMessage(commonRecord *models.EnvelopeEventCommon, postItems []string) error {
	if len(postItems) < 3 {
		return nil
	}

	var p fastjson.Parser
	v, err := p.Parse(postItems[models.EnvelopePostItemMessage])
	if err != nil {
		return nil
	}

	if v.Exists("message") {
		commonRecord.Message = string(v.Get("message").GetStringBytes())
	}
	if v.Exists("level") {
		commonRecord.Level = string(v.Get("level").GetStringBytes())
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

// formatValue is no longer used but kept for reference.
// Use string(val.GetStringBytes()) to correctly decode unicode escapes.
func formatValue(val string) string {
	return strings.ReplaceAll(val, "\"", "")
}
