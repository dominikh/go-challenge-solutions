package drum

import (
	"encoding/binary"
	"io"
	"os"
)

type stickyErrWriter struct {
	w   io.Writer
	err error
}

func (w *stickyErrWriter) write(b []byte) {
	if w.err != nil {
		return
	}
	_, w.err = w.w.Write(b)
}

func (w *stickyErrWriter) bwrite(bo binary.ByteOrder, v interface{}) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.w, bo, v)
}

func Encode(w io.Writer, p *Pattern) error {
	sw := stickyErrWriter{w: w}

	size := 32 + 4
	size += len(p.Tracks) * (5 + 16)
	for _, t := range p.Tracks {
		size += len(t.Name)
	}

	sw.write(Magic)
	sw.bwrite(binary.BigEndian, int64(size))
	var version [32]byte
	if len(p.Version) > 32 {
		// FIXME return error
	}
	copy(version[:], p.Version)
	sw.bwrite(binary.LittleEndian, version)
	sw.bwrite(binary.LittleEndian, p.BPM)

	type header struct {
		ID         int32
		NameLength byte
	}
	for _, t := range p.Tracks {
		if len(t.Name) > 0xFF {
			// FIXME return error
		}
		h := header{
			ID:         int32(t.ID),
			NameLength: byte(len(t.Name)),
		}
		sw.bwrite(binary.LittleEndian, h)
		sw.write([]byte(t.Name))
		for i := 0; i < 16; i++ {
			if t.Step(i) {
				sw.write([]byte{1})
			} else {
				sw.write([]byte{0})
			}
		}
	}
	return sw.err
}

func EncodeFile(name string, p *Pattern) (err error) {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		err1 := f.Close()
		if err == nil {
			err = err1
		}
	}()
	return Encode(f, p)
}
