package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

/*
With the exception of the file size, which is encoded in big endian,
all values are encoded in little endian.

The SPLICE header is 50 bytes long and consists of the following fields:

File identifier: "SPLICE"
File size: int64 (big endian)
Version: [32]byte
BPM: float32

Tracks are of variable length, consisting of a 5 byte header, a
variable length name, and 16 bytes for the steps.

Track ID: int32
Length of name: byte
Name: [length of name]byte
Steps: [16]byte
*/

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

	limited := io.LimitReader(r, header.Size).(*io.LimitedReader)
	r = limited
	err = binary.Read(r, binary.LittleEndian, &p)
	if err != nil {
		return nil, err
	}

	var tracks []Track
	for limited.N > 0 {
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
