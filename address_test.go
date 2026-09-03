package benedict

import (
	"errors"
	"testing"
)

func TestNewAddress(t *testing.T) {
	cases := []struct {
		address  string
		name     string
		expected *Address
	}{
		{
			"addr@example.com",
			"John Smith",
			&Address{address: "addr@example.com", name: "John Smith"},
		},
		{
			"addr@example.com",
			"山田 太郎",
			&Address{address: "addr@example.com", name: "山田 太郎"},
		},
		{
			"addr@example.com",
			"",
			&Address{address: "addr@example.com", name: ""},
		},
		{ // Not Mail: newAddress() itself performs no validation, only parse() does
			"addratexample.com",
			"",
			&Address{address: "addratexample.com", name: ""},
		},
		{ // line breaks must be stripped from both address and name
			"addr@example.com\r\n",
			"John\r\nSmith",
			&Address{address: "addr@example.com", name: "JohnSmith"},
		},
	}

	for i, v := range cases {
		a := newAddress(v.address, v.name)
		if a.address != v.expected.address {
			t.Errorf("[Case%d] Address Result:%v Expected:%v", i, a.address, v.expected.address)
		}
		if a.name != v.expected.name {
			t.Errorf("[Case%d] Name Result:%v Expected:%v", i, a.name, v.expected.name)
		}
	}
}

func TestAddressParse(t *testing.T) {
	cases := []struct {
		address    string
		name       string
		expAddress string
		expName    string
		expected   error
	}{
		{
			"john@example.com",
			"John",
			"john@example.com",
			"John",
			nil,
		},
		{
			"john@example.com",
			"",
			"john@example.com",
			"",
			nil,
		},
		{
			"0123456",
			"Test",
			"",
			"",
			errors.New("mail: missing @ in addr-spec"),
		},
	}
	for i, c := range cases {
		a := newAddress(c.address, c.name)
		err := a.parse()
		if a.address != c.expAddress {
			t.Errorf(`[Case%d] Address Result: %v Expected: %v`, i, a.address, c.expAddress)
		}
		if a.name != c.expName {
			t.Errorf(`[Case%d] Name Result: %v Expected: %v`, i, a.name, c.expName)
		}
		if c.expected != nil && err.Error() != c.expected.Error() {
			t.Errorf(`[Case%d] Error Result: %v Expected: %v`, i, err, c.expected)
		}
	}
}

func TestAddressAngle(t *testing.T) {
	cases := []struct {
		address string
		expect  string
	}{
		{
			"john@example.com",
			"<john@example.com>",
		},
		{
			"janeatexample.com",
			"",
		},
		{
			"",
			"",
		},
	}

	for k, v := range cases {
		a := newAddress(v.address, "")
		_ = a.parse()
		if a.Angle() != v.expect {
			t.Errorf("[Case%d] Result:%v Expected:%v", k, a.Angle(), v.expect)
		}
	}
}

func TestAddressString(t *testing.T) {
	cases := []struct {
		name     string
		address  string
		expected string
	}{
		{
			"John Smith",
			"john@example.com",
			"John Smith <john@example.com>",
		},
		{
			"",
			"jane@example.com",
			"<jane@example.com>",
		},
		{
			"山田 太郎",
			"taro@example.com",
			"=?UTF-8?B?5bGx55SwIOWkqumDjg==?= <taro@example.com>",
		},
		{
			"",
			"",
			"",
		},
	}

	for i, c := range cases {
		a := newAddress(c.address, c.name)
		if a.String() != c.expected {
			t.Errorf("[Case%d] Result: %v Expected: %v", i, a.String(), c.expected)
		}
	}
}
