package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// Database-mapped JSON column types.
// MySQL returns JSON columns as []byte; these wrappers decode them into typed
// slices (and re-encode on write), playing the role of MyBatis TypeHandlers.

type StringList []string

func (s *StringList) Scan(value interface{}) error {
	return scanJSON(value, s)
}

func (s StringList) Value() (driver.Value, error) {
	return marshalJSON(s)
}

type WaypointList []Waypoint

func (w *WaypointList) Scan(value interface{}) error {
	return scanJSON(value, w)
}

func (w WaypointList) Value() (driver.Value, error) {
	return marshalJSON(w)
}

type CoordinateList []Coordinate

func (c *CoordinateList) Scan(value interface{}) error {
	return scanJSON(value, c)
}

func (c CoordinateList) Value() (driver.Value, error) {
	return marshalJSON(c)
}

func scanJSON(value interface{}, target interface{}) error {
	if value == nil {
		return nil
	}
	raw, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected JSON bytes, got %T", value)
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return json.Unmarshal(raw, target)
}

func marshalJSON(v interface{}) (driver.Value, error) {
	// A nil/empty slice marshals to `null`, which the API contract forbids; write [] instead.
	if isEmptySlice(v) {
		return "[]", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, errors.New("failed to marshal JSON column")
	}
	return string(b), nil
}

func isEmptySlice(v interface{}) bool {
	switch s := v.(type) {
	case StringList:
		return len(s) == 0
	case WaypointList:
		return len(s) == 0
	case CoordinateList:
		return len(s) == 0
	}
	return false
}
