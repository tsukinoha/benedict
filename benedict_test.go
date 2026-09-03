package benedict

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"
)

func TestEencode(t *testing.T) {
	cases := []struct {
		param   []byte
		flg     bool
		exptect []byte
	}{
		{
			[]byte("Mike Davice"),
			false,
			[]byte("Mike Davice"),
		},
		{
			[]byte("山田 太郎"),
			false,
			[]byte("5bGx55SwIOWkqumDjg=="),
		},
		{
			[]byte("Mike Davice"),
			true,
			[]byte("Mike Davice"),
		},
		{
			[]byte("山田 太郎"),
			true,
			[]byte("=?UTF-8?B?5bGx55SwIOWkqumDjg==?="),
		},
	}
	for i, c := range cases {
		result := encodeMime(c.param, c.flg)
		if !bytes.Equal(c.exptect, result) {
			t.Errorf(`[Case%d] Expect: %s,  Result: %s`, i+1, c.exptect, result)
		}
	}
}

func TestEncodeString(t *testing.T) {
	cases := []struct {
		param   string
		flg     bool
		exptect string
	}{
		{
			"Mike Davice",
			false,
			"Mike Davice",
		},
		{
			"山田 太郎",
			false,
			"5bGx55SwIOWkqumDjg==",
		},
		{
			"Mike Davice",
			true,
			"Mike Davice",
		},
		{
			"山田 太郎",
			true,
			"=?UTF-8?B?5bGx55SwIOWkqumDjg==?=",
		},
	}

	for i, c := range cases {
		result := encodeMimeString(c.param, c.flg)
		if c.exptect != result {
			t.Errorf(`[Case%d] Expect: %s,  Result: %s`, i+1, c.exptect, result)
		}
	}
}

func TestBorder(t *testing.T) {
	cases := []struct {
		length int
		want   int
	}{
		{24, 24},
		{8, 8},
		{0, 24},  // length < 1 falls back to the default of 24
		{-5, 24}, // length < 1 falls back to the default of 24
	}
	for i, c := range cases {
		b := border(c.length)
		re := regexp.MustCompile(fmt.Sprintf(`^-{12}[0-9a-zA-Z]{%d}$`, c.want))
		if !re.MatchString(b) {
			t.Errorf(`[Case%d] Invalid Border %s`, i, b)
		}
	}
}

func TestRemoveBreak(t *testing.T) {
	cases := []struct {
		param    string
		expected string
	}{
		{
			"no break",
			"no break",
		},
		{
			"line1\r\nline2",
			"line1line2",
		},
		{
			"line1\nline2\r\n",
			"line1line2",
		},
		{
			"\r\r\n\n",
			"",
		},
		{
			"",
			"",
		},
	}
	for i, c := range cases {
		result := removeBreak(c.param)
		if result != c.expected {
			t.Errorf(`[Case%d] Expect: %q, Result: %q`, i, c.expected, result)
		}
	}
}

func TestIsLocalhost(t *testing.T) {
	cases := []struct {
		name     string
		expected bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"example.com", false},
		{"", false},
	}
	for i, c := range cases {
		result := isLocalhost(c.name)
		if result != c.expected {
			t.Errorf(`[Case%d] Name:%s Result:%v Expected:%v`, i, c.name, result, c.expected)
		}
	}
}
