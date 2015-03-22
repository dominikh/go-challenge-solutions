package drum

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"path"
	"testing"
)

func TestEncodePattern(t *testing.T) {
	p := &Pattern{
		Version: "0.808-alpha",
		BPM:     120,
		Tracks: []Track{
			{0, "kick", []bool{true, false, false, false,
				true, false, false, false,
				true, false, false, false,
				true, false, false, false}},
			{1, "snare", []bool{false, false, false,
				false, true, false, false,
				false, false, false, false,
				false, true, false, false, false}},
			{2, "clap", []bool{false, false, false, false,
				true, false, true, false,
				false, false, false, false,
				false, false, false, false}},
			{3, "hh-open", []bool{false, false, true, false,
				false, false, true, false,
				true, false, true, false,
				false, false, true, false}},
			{4, "hh-close", []bool{true, false, false, false,
				true, false, false, false,
				false, false, false, false,
				true, false, false, true}},
			{5, "cowbell", []bool{false, false, false, false,
				false, false, false, false,
				false, false, true, false,
				false, false, false, false}},
		},
	}
	w := &bytes.Buffer{}

	err := Encode(w, p)
	if err != nil {
		t.Fatalf("Got error %q", err)
	}

	expected, err := ioutil.ReadFile(path.Join("fixtures", "pattern_1.splice"))
	if err != nil {
		t.Fatalf("Got error %q", err)
	}

	b := w.Bytes()
	if !bytes.Equal(expected, b) {
		t.Error("failed encoding")
		t.Logf("Got:\n%s", hex.Dump(b))
		t.Logf("Expected:\n%s", hex.Dump(expected))
	}
}
