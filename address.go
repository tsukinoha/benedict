package benedict

import (
	"fmt"
	"net/mail"
)

type (
	Address struct {
		address string
		name    string
	}
)

func newAddress(address, name string) *Address {
	a := new(Address)
	a.address = removeBreak(address)
	a.name = removeBreak(name)
	return a
}

func (a *Address) parse() error {
	_, err := mail.ParseAddress(fmt.Sprintf("%s <%s>", a.name, a.address))
	if err != nil {
		a.address = ""
		a.name = ""
		return err
	}
	return nil
}

func (a *Address) Angle() string {
	if a.address == "" {
		return ""
	} else {
		return fmt.Sprintf("<%s>", a.address)
	}
}

func (a *Address) String() string {
	if a.address == "" {
		return ""
	} else if len(a.name) > 0 {
		return fmt.Sprintf("%s %s", encodeMimeString(a.name, true), a.Angle())
	} else {
		return a.Angle()
	}
}
