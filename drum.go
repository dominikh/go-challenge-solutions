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
	ID   int
	Name string

	// We're trading speed for memory. Instead of storing 16 bools (==
	// 16 bytes), which can be indexed directly, we store a single
	// 16-bit int and use bit shifting and masks. The idea is that an
	// old drum machine is bottlenecked by memory more than it is by
	// CPU, that cache lines are tiny and that we're better off
	// spending a cycle more on bit shifts than on fetching 16 bytes
	// of memory. This is especially true if we want to store a lot of
	// tracks at once.
	steps uint16
}

func (t Track) SetStep(i int, b bool) {
	if b {
		t.steps |= 1 << uint(i)
	} else {
		t.steps &^= 1 << uint(i)
	}
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
