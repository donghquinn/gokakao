package kakao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Time is a timestamp in a Kakao response. It decodes RFC 3339 strings
// with or without fractional seconds, ISO 8601 strings with a "+0000" style
// offset, local times without an offset (read as KST), and JSON numbers as
// Unix seconds; an empty string or null decodes to the zero time.
type Time struct {
	time.Time
}

// UnmarshalJSON implements [json.Unmarshaler].
func (t *Time) UnmarshalJSON(b []byte) error {
	var s string
	if bytes.Equal(b, []byte("null")) {
		t.Time = time.Time{}
		return nil
	}
	if len(b) > 0 && b[0] != '"' {
		sec, err := strconv.ParseInt(string(b), 10, 64)
		if err != nil {
			return fmt.Errorf("kakao: time: %w", err)
		}
		t.Time = time.Unix(sec, 0)
		return nil
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("kakao: time: %w", err)
	}
	if s == "" {
		t.Time = time.Time{}
		return nil
	}
	v, err := parseTime(s)
	if err != nil {
		return fmt.Errorf("kakao: time: %w", err)
	}
	t.Time = v
	return nil
}

// kst is Korea Standard Time, assumed for timestamps without an offset.
var kst = time.FixedZone("KST", 9*60*60)

// parseTime parses the timestamp formats Kakao APIs use: RFC 3339, ISO
// 8601 with a "+0000" style offset, and local times without an offset,
// which are read as KST.
func parseTime(s string) (time.Time, error) {
	v, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return v, nil
	}
	if v, err2 := time.Parse("2006-01-02T15:04:05.999999999-0700", s); err2 == nil {
		return v, nil
	}
	for _, layout := range []string{"2006-01-02T15:04:05.999999999", "2006-01-02 15:04:05"} {
		if v, err2 := time.ParseInLocation(layout, s, kst); err2 == nil {
			return v, nil
		}
	}
	return time.Time{}, err
}

// Float is a number in a Kakao response that may be encoded either as a
// JSON number or as a string, such as coordinates ("127.06") and distances
// ("418"). An empty string or null decodes to 0.
type Float float64

// UnmarshalJSON implements [json.Unmarshaler].
func (f *Float) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*f = 0
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return fmt.Errorf("kakao: number: %w", err)
		}
		if s == "" {
			*f = 0
			return nil
		}
		b = []byte(s)
	}
	v, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return fmt.Errorf("kakao: number: %w", err)
	}
	*f = Float(v)
	return nil
}

// Int64 is an integer in a Kakao response that may be encoded either as a
// JSON number or as a string, such as IDs in the advertising APIs. An empty
// string or null decodes to 0. It encodes as a JSON number.
type Int64 int64

// UnmarshalJSON implements [json.Unmarshaler].
func (n *Int64) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*n = 0
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return fmt.Errorf("kakao: integer: %w", err)
		}
		if s == "" {
			*n = 0
			return nil
		}
		b = []byte(s)
	}
	v, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return fmt.Errorf("kakao: integer: %w", err)
	}
	*n = Int64(v)
	return nil
}

// ID is an identifier that Kakao encodes as a JSON string in some
// responses and as a number in others. It is always a string in Go and
// encodes as a JSON string.
type ID string

// UnmarshalJSON implements [json.Unmarshaler].
func (id *ID) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*id = ""
		return nil
	}
	if len(b) > 0 && b[0] != '"' {
		var n json.Number
		if err := json.Unmarshal(b, &n); err != nil {
			return fmt.Errorf("kakao: ID: %w", err)
		}
		if _, err := n.Int64(); err != nil {
			return fmt.Errorf("kakao: ID: %w", err)
		}
		*id = ID(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("kakao: ID: %w", err)
	}
	*id = ID(s)
	return nil
}

// YesNo is a "Y"/"N" flag in a Kakao response. "Y" decodes to true and
// anything else to false.
type YesNo bool

// UnmarshalJSON implements [json.Unmarshaler].
func (y *YesNo) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*y = false
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("kakao: Y/N flag: %w", err)
	}
	*y = s == "Y"
	return nil
}

// MarshalJSON implements [json.Marshaler], encoding "Y" or "N".
func (y YesNo) MarshalJSON() ([]byte, error) {
	if y {
		return []byte(`"Y"`), nil
	}
	return []byte(`"N"`), nil
}

// formatFloat formats f without exponent or trailing zeros.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
