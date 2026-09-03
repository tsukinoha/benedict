package benedict

import (
	"io"
)

type Att struct {
	r    io.Reader
	name string
}

func NewAtt(r io.Reader, name string) (*Att, error) {
	a := new(Att)
	a.name = name
	a.r = r
	return a, nil
}

func (a *Att) Name() string {
	return a.name
}

func (a *Att) Encode() string {
	return ""
}
