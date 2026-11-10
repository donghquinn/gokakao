package kakao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Time is a timestamp in a Kakao response. It decodes ISO 8601 / RFC 3339
// values with or without fractional seconds; an empty string or null
// decodes to the zero time.
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
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("kakao: time: %w", err)
	}
	if s == "" {
		t.Time = time.Time{}
		return nil
	}
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("kakao: time: %w", err)
	}
	t.Time = v
	return nil
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
