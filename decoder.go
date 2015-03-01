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
var ErrInvalidHeader = errors.New("input is missing valid SPLICE header")

func Decode(r io.Reader) (*Pattern, error) {
	var header struct {
		Magic [6]byte
		Size  int64
	}
	var p struct {
		Version [32]byte
		BPM     float32
	}

	// TODO(dominikh): Switching between little endian and big endian
	// in the same file format is weird, but otherwise there'd only be
	// one byte for the file size, which seems awfully small, so let's
	// assume the file format is weird.
	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(header.Magic[:], Magic) {
		return nil, ErrInvalidHeader
	}
	if header.Size < 0 {
		return nil, ErrInvalidHeader
	}

	limited := io.LimitReader(r, int64(header.Size)).(*io.LimitedReader)
	r = limited
	err = binary.Read(r, binary.LittleEndian, &p)
	if err != nil {
		return nil, err
	}

	var tracks []Track
	for {
		if limited.N == 0 {
			break
		}
		track, err := readTrack(r)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	version := p.Version[:]
	end := bytes.IndexByte(version, 0)
	if end > -1 {
		version = version[:end]
	}
	return &Pattern{Version: string(version), BPM: p.BPM, Tracks: tracks}, nil
}

func readTrack(r io.Reader) (Track, error) {
	var header struct {
		ID         int32
		NameLength byte
	}
	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return Track{}, err
	}
	name := make([]byte, header.NameLength)
	_, err = io.ReadFull(r, name)
	if err != nil {
		return Track{}, err
	}

	var steps [16]byte
	err = binary.Read(r, binary.LittleEndian, steps[:])
	var stepFlag uint16
	for i, st := range steps {
		if st > 0 {
			stepFlag |= 1 << uint(i)
		}
	}
	return Track{ID: int(header.ID), Name: string(name), steps: stepFlag}, err
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
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
	steps uint16
}

func (t Track) Step(i int) bool {
	return (t.steps & (1 << uint(i))) > 0
}

func (t Track) Steps() []bool {
	steps := make([]bool, 16)
	for i := 0; i < 16; i++ {
		steps[i] = t.Step(i)
	}
	return steps
}

func (t Track) formatSteps() string {
	s := make([]byte, 16)
	for i := 0; i < 16; i++ {
		if t.Step(i) {
			s[i] = 'x'
		} else {
			s[i] = '-'
		}
	}
	return string(s)
}

func (t *Track) String() string {
	st := t.formatSteps()

	return fmt.Sprintf("(%d) %s\t|%s|%s|%s|%s|", t.ID, t.Name,
		st[0:4], st[4:8], st[8:12], st[12:16])
}
