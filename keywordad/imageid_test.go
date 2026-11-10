package keywordad

import (
	"encoding/json"
	"testing"
)

func TestImageID_JSON(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]ImageID{`8888888883`: "8888888883", `"8888888881"`: "8888888881", `null`: "", `""`: ""} {
		var id ImageID
		if err := json.Unmarshal([]byte(in), &id); err != nil || id != want {
			t.Errorf("%s = %q, %v", in, id, err)
		}
	}
	for _, in := range []string{`1.5`, `true`} {
		var id ImageID
		if err := json.Unmarshal([]byte(in), &id); err == nil {
			t.Errorf("%s: want error", in)
		}
	}
	if b, _ := json.Marshal(ThumbnailAsset{ImageID: "1"}); string(b) != `{"imageId":"1"}` {
		t.Errorf("Marshal = %s", b)
	}
}
