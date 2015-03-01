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
			{0, "kick", 0x1111},
			{1, "snare", 0x1010},
			{2, "clap", 0x0050},
			{3, "hh-open", 0x4544},
			{4, "hh-close", 0x9011},
			{5, "cowbell", 0x0400},
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
