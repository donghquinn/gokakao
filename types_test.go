package kakao

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTime_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	kst := time.FixedZone("", 9*60*60)
	tests := []struct {
		name    string
		in      string
		want    time.Time
		wantErr bool
	}{
		{"fractional seconds", `"2017-06-14T00:00:00.000+09:00"`, time.Date(2017, 6, 14, 0, 0, 0, 0, kst), false},
		{"no fraction", `"2017-05-06T00:36:45+09:00"`, time.Date(2017, 5, 6, 0, 36, 45, 0, kst), false},
		{"empty string", `""`, time.Time{}, false},
		{"null", `null`, time.Time{}, false},
		{"unix seconds", `1658214426`, time.Unix(1658214426, 0), false},
		{"not a time", `true`, time.Time{}, true},
		{"fractional number", `1.5`, time.Time{}, true},
		{"bad format", `"2017/06/14"`, time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Time
			err := json.Unmarshal([]byte(tt.in), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got.Time, tt.want)
			}
		})
	}
}

func TestFloat_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    Float
		wantErr bool
	}{
		{"number", `127.10459896729914`, 127.10459896729914, false},
		{"string", `"37.51207412593136"`, 37.51207412593136, false},
		{"integer string", `"418"`, 418, false},
		{"empty string", `""`, 0, false},
		{"null", `null`, 0, false},
		{"bad string", `"abc"`, 0, true},
		{"bool", `true`, 0, true},
		{"unterminated string", `"12`, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Float
			err := got.UnmarshalJSON([]byte(tt.in))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYesNo_JSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    YesNo
		wantErr bool
	}{
		{`"Y"`, true, false},
		{`"N"`, false, false},
		{`""`, false, false},
		{`null`, false, false},
		{`1`, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			var got YesNo
			err := json.Unmarshal([]byte(tt.in), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
	for v, want := range map[YesNo]string{true: `"Y"`, false: `"N"`} {
		b, err := json.Marshal(v)
		if err != nil || string(b) != want {
			t.Errorf("Marshal(%v) = %s, %v; want %s", v, b, err, want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	t.Parallel()
	for in, want := range map[float64]string{
		127.1086228:        "127.1086228",
		37:                 "37",
		-4388.879299157299: "-4388.879299157299",
		0.00000123:         "0.00000123",
		160710.37729270622: "160710.37729270622",
	} {
		if got := formatFloat(in); got != want {
			t.Errorf("formatFloat(%v) = %q, want %q", in, got, want)
		}
	}
}

func FuzzLenientTypes(f *testing.F) {
	for _, s := range []string{`"2017-06-14T00:00:00.000+09:00"`, `""`, `null`, `127.5`, `"37.1"`, `"Y"`, `"N"`, `"`, `{}`, `[1]`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(_ *testing.T, b []byte) {
		var tm Time
		_ = tm.UnmarshalJSON(b)
		var fl Float
		_ = fl.UnmarshalJSON(b)
		var yn YesNo
		_ = yn.UnmarshalJSON(b)
	})
}
