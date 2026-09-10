package alpaca

import (
	"encoding/json"
	"fmt"
	"time"

	"bar_sequence_lab/generator/internal/types"
)

type control struct {
	Type    string   `json:"T"`
	Message string   `json:"msg"`
	Code    int      `json:"code"`
	Bars    []string `json:"bars"`
}

func decodeMessageArray(raw []byte) ([]json.RawMessage, error) {
	var messages []json.RawMessage
	if err := json.Unmarshal(raw, &messages); err != nil {
		return nil, fmt.Errorf("decode alpaca array: %w", err)
	}
	return messages, nil
}

func decodeControl(raw json.RawMessage) (control, error) {
	fields, err := rawObject(raw)
	if err != nil {
		return control{}, err
	}
	var c control
	c.Type = rawString(fields, "T")
	c.Message = rawString(fields, "msg")
	c.Code = rawInt(fields, "code")
	if bars, ok := fields["bars"]; ok {
		_ = json.Unmarshal(bars, &c.Bars)
	}
	return c, nil
}

func observationFromRaw(raw json.RawMessage, sourceID string, receipt time.Time) (types.Observation, error) {
	fields, err := rawObject(raw)
	if err != nil {
		return types.Observation{}, err
	}
	open, err := requiredFloat(fields, "o")
	if err != nil {
		return types.Observation{}, err
	}
	high, err := requiredFloat(fields, "h")
	if err != nil {
		return types.Observation{}, err
	}
	low, err := requiredFloat(fields, "l")
	if err != nil {
		return types.Observation{}, err
	}
	closePx, err := requiredFloat(fields, "c")
	if err != nil {
		return types.Observation{}, err
	}
	volume, err := requiredUint(fields, "v")
	if err != nil {
		return types.Observation{}, err
	}
	symbol := rawString(fields, "S")
	if symbol == "" {
		return types.Observation{}, fmt.Errorf("alpaca bar missing symbol")
	}
	ts := rawString(fields, "t")
	if ts == "" {
		return types.Observation{}, fmt.Errorf("alpaca bar missing timestamp")
	}
	start, err := parseAlpacaTime(ts)
	if err != nil {
		return types.Observation{}, fmt.Errorf("parse alpaca timestamp %q: %w", ts, err)
	}
	if err := validateOHLC(open, high, low, closePx); err != nil {
		return types.Observation{}, err
	}
	msgType := rawString(fields, "T")
	if msgType != "b" && msgType != "u" {
		return types.Observation{}, fmt.Errorf("unsupported alpaca message type %q", msgType)
	}
	trades, _ := optionalUint32(fields, "n")
	return types.Observation{
		Symbol:              symbol,
		SourceEventTime:     start,
		SourceTimestampText: ts,
		ReceivedTime:        receipt,
		Interval:            types.Interval1Min,
		Open:                open,
		High:                high,
		Low:                 low,
		Close:               closePx,
		Volume:              volume,
		EventCount:          trades,
		SourceID:            sourceID,
		AlpacaMessageType:   msgType,
		PayloadHash:         types.HashRaw(raw),
	}, nil
}

func rawObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func rawString(fields map[string]json.RawMessage, key string) string {
	value, ok := fields[key]
	if !ok {
		return ""
	}
	var parsed string
	if err := json.Unmarshal(value, &parsed); err != nil {
		return ""
	}
	return parsed
}

func rawInt(fields map[string]json.RawMessage, key string) int {
	value, ok := fields[key]
	if !ok {
		return 0
	}
	var parsed int
	if err := json.Unmarshal(value, &parsed); err != nil {
		return 0
	}
	return parsed
}

func requiredFloat(fields map[string]json.RawMessage, key string) (float64, error) {
	value, ok := fields[key]
	if !ok {
		return 0, fmt.Errorf("alpaca bar missing %s", key)
	}
	var parsed float64
	if err := json.Unmarshal(value, &parsed); err != nil {
		return 0, fmt.Errorf("alpaca bar unreadable %s", key)
	}
	return parsed, nil
}

func requiredUint(fields map[string]json.RawMessage, key string) (uint64, error) {
	value, ok := fields[key]
	if !ok {
		return 0, fmt.Errorf("alpaca bar missing %s", key)
	}
	var n json.Number
	if err := json.Unmarshal(value, &n); err == nil {
		parsed, err := n.Int64()
		if err != nil {
			f, ferr := n.Float64()
			if ferr != nil || f < 0 {
				return 0, fmt.Errorf("alpaca bar unreadable %s", key)
			}
			return uint64(f), nil
		}
		if parsed < 0 {
			return 0, fmt.Errorf("negative volume")
		}
		return uint64(parsed), nil
	}
	var asString string
	if err := json.Unmarshal(value, &asString); err == nil {
		n = json.Number(asString)
		parsed, err := n.Int64()
		if err != nil || parsed < 0 {
			return 0, fmt.Errorf("alpaca bar unreadable %s", key)
		}
		return uint64(parsed), nil
	}
	return 0, fmt.Errorf("alpaca bar unreadable %s", key)
}

func optionalUint32(fields map[string]json.RawMessage, key string) (uint32, error) {
	if _, ok := fields[key]; !ok {
		return 0, nil
	}
	n, err := requiredUint(fields, key)
	return uint32(n), err
}

func validateOHLC(open, high, low, close float64) error {
	for _, value := range []float64{open, high, low, close} {
		if value <= 0 || value != value {
			return fmt.Errorf("alpaca bar incomplete ohlc")
		}
	}
	if high < low || high < open || high < close || low > open || low > close {
		return fmt.Errorf("alpaca bar inconsistent ohlc")
	}
	return nil
}

func parseAlpacaTime(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.UTC(), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}
	if parsed, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return parsed.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp %q", value)
}
