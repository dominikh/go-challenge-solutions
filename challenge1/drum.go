// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import "fmt"

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
	Steps []bool
}

func (t Track) formatSteps() string {
	s := make([]byte, 16)
	for i, step := range t.Steps {
		if step {
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
