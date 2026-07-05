package service

import (
	"fmt"

	"github.com/GORATOR/backend/internal/models"
	"github.com/valyala/fastjson"
)

func ParseTransaction(postItems []string) (*models.Transaction, error) {
	payload, ok := models.FindEnvelopeItemPayload(postItems, "transaction")
	if !ok {
		return nil, fmt.Errorf("no transaction item found in envelope")
	}

	var p fastjson.Parser
	v, err := p.Parse(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction payload: %w", err)
	}

	tx := &models.Transaction{
		EventId:        string(v.GetStringBytes("event_id")),
		Name:           string(v.GetStringBytes("transaction")),
		StartTimestamp: extractTimestampValue(v, "start_timestamp"),
		EndTimestamp:   extractTimestampValue(v, "timestamp"),
	}

	// Fallback: event_id from envelope header
	if tx.EventId == "" {
		header, err := p.Parse(postItems[models.EnvelopePostItemCommon])
		if err == nil {
			tx.EventId = string(header.GetStringBytes("event_id"))
		}
	}

	// trace_id, span_id, op, status live in contexts.trace
	if traceCtx := v.Get("contexts", "trace"); traceCtx != nil {
		tx.TraceId = string(traceCtx.GetStringBytes("trace_id"))
		tx.SpanId = string(traceCtx.GetStringBytes("span_id"))
		tx.ParentSpanId = string(traceCtx.GetStringBytes("parent_span_id"))
		tx.Op = string(traceCtx.GetStringBytes("op"))
		tx.Status = string(traceCtx.GetStringBytes("status"))
	}

	if ctxObj := v.Get("contexts"); ctxObj != nil {
		s := ctxObj.String()
		tx.Contexts = &s
	}

	if tagsObj := v.Get("tags"); tagsObj != nil {
		s := tagsObj.String()
		tx.Tags = &s
	}

	for _, spanVal := range v.GetArray("spans") {
		span := models.Span{
			SpanId:         string(spanVal.GetStringBytes("span_id")),
			ParentSpanId:   string(spanVal.GetStringBytes("parent_span_id")),
			TraceId:        tx.TraceId,
			Op:             string(spanVal.GetStringBytes("op")),
			Description:    string(spanVal.GetStringBytes("description")),
			Status:         string(spanVal.GetStringBytes("status")),
			StartTimestamp: extractTimestampValue(spanVal, "start_timestamp"),
			EndTimestamp:   extractTimestampValue(spanVal, "timestamp"),
		}
		if dataObj := spanVal.Get("data"); dataObj != nil {
			s := dataObj.String()
			span.Data = &s
		}
		tx.Spans = append(tx.Spans, span)
	}

	return tx, nil
}

// extractTimestampValue handles both float (Unix) and string (ISO 8601) formats sent by Sentry SDKs.
func extractTimestampValue(v *fastjson.Value, key string) string {
	val := v.Get(key)
	if val == nil {
		return ""
	}
	if val.Type() == fastjson.TypeString {
		return string(val.GetStringBytes())
	}
	return val.String()
}
