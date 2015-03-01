package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

var Magic = []byte("SPLICE")
var ErrMissingHeader = errors.New("input is missing valid SPLICE header")

func Decode(r io.Reader) (*Pattern, error) {
	p := &pattern{}
	var magic [6]byte
	_, err := io.ReadFull(r, magic[:])
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(magic[:], Magic) {
		return nil, ErrMissingHeader
	}
	var length int64
	// TODO(dominikh): Switching between little endian and big endian
	// in the same file format is weird, but otherwise there'd only be
	// one byte for the file size, which seems awfully small, so let's
	// assume the file format is weird.
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		println(2)
		return nil, err
	}
	// TODO error if length is negative
	// TODO error if remaining length > 0 after we're done (trailing data is evil)
	limited := io.LimitReader(r, int64(length)).(*io.LimitedReader)
	r = limited
	err = binary.Read(r, binary.LittleEndian, p)
	if err != nil {
		println(3)
		return nil, err
	}

	var tracks []Track
	// TODO sticky error
	for {
		var id int32
		err = binary.Read(r, binary.LittleEndian, &id)
		if err != nil {
			if err == io.EOF {
				break
				// FIXME check that we read zero bytes
			}
			println(4)
			return nil, err
		}
		var n byte
		err = binary.Read(r, binary.LittleEndian, &n)
		if err != nil {
			println(5)
			return nil, err
		}
		b := make([]byte, n)
		_, err = io.ReadFull(r, b)
		if err != nil {
			println(6)
			return nil, err
		}

		var ticks [16]byte
		// FIXME rename `ticks`
		err = binary.Read(r, binary.LittleEndian, ticks[:])
		if err != nil {
			println(7)
			return nil, err
		}

		tracks = append(tracks, Track{ID: int(id), Name: string(b), Ticks: ticks})
	}
	version := p.Version[:]
	end := bytes.IndexByte(version, 0)
	if end > -1 {
		version = version[:end]
	}
	return &Pattern{Version: string(version), BPM: p.BPM, Tracks: tracks}, nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Decode(f)
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type pattern struct {
	// FIXME it's probably null terminated, not fixed length. or the length is encoded somewhere else
	Version [32]byte
	BPM     float32
}

type Pattern struct {
	Version string
	BPM     float32
	Tracks  []Track
}

func (p *Pattern) String() string {
	s := fmt.Sprintf(`Saved with HW Version: %s
Tempo: %g
`, p.Version, p.BPM)

	for _, t := range p.Tracks {
		s += t.String() + "\n"
	}
	return s
}

type Track struct {
	ID    int
	Name  string
	Ticks [16]byte
}

func formatTicks(t []byte) string {
	s := ""
	for _, b := range t {
		if b == 1 {
			s += "x"
		} else {
			s += "-"
		}
	}
	return s
}

func (t *Track) String() string {
	ti := t.Ticks
	ticks := "|" + formatTicks(ti[0:4]) +
		"|" + formatTicks(ti[4:8]) +
		"|" + formatTicks(ti[8:12]) +
		"|" + formatTicks(ti[12:16]) + "|"
	return fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, ticks)
}
