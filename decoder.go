package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func Decode(r io.Reader) (*Pattern, error) {
	p := &pattern{}
	// 6 bytes SPLICE header, 7 bytes no idea
	// TODO actually do read SPLICE header, for verification
	// 1 byte is our file size, 7 bytes are unknown… maybe it's an
	// int64 in big endian for the file size? but why would endianness
	// switch in the middle of the file?
	var scratch [13]byte
	_, err := io.ReadFull(r, scratch[:])
	if err != nil {
		println(1)
		return nil, err
	}
	var length byte
	err = binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		println(2)
		return nil, err
	}
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

// ${byte track name length}${string track name}16x${boolean}
// ................

// c5, 8f, c5, 93, 57

// 00000000  53 50 4c 49 43 45 00 00  00 00 00 00 00 c5 30 2e  |SPLICE........0.|
// 00000010  38 30 38 2d 61 6c 70 68  61 00 00 00 00 00 00 00  |808-alpha.......|
// 00000020  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
// 00000030  f0 42 00 00 00 00 04 6b  69 63 6b 01 00 00 00 01  |.B.....kick.....|
// 00000040  00 00 00 01 00 00 00 01  00 00 00 01 00 00 00 05  |................|
// 00000050  73 6e 61 72 65 00 00 00  00 01 00 00 00 00 00 00  |snare...........|
// 00000060  00 01 00 00 00 02 00 00  00 04 63 6c 61 70 00 00  |..........clap..|
// 00000070  00 00 01 00 01 00 00 00  00 00 00 00 00 00 03 00  |................|
// 00000080  00 00 07 68 68 2d 6f 70  65 6e 00 00 01 00 00 00  |...hh-open......|
// 00000090  01 00 01 00 01 00 00 00  01 00 04 00 00 00 08 68  |...............h|
// 000000a0  68 2d 63 6c 6f 73 65 01  00 00 00 01 00 00 00 00  |h-close.........|
// 000000b0  00 00 00 01 00 00 01 05  00 00 00 07 63 6f 77 62  |............cowb|
// 000000c0  65 6c 6c 00 00 00 00 00  00 00 00 00 00 01 00 00  |ell.............|
// 000000d0  00 00 00                                          |...|
// 000000d3
