package benedict

import (
	"strings"
	"testing"
)

func TestNewDsn(t *testing.T) {
	cases := []struct {
		dsn string
	}{
		{
			"smtp://example.com:25",
		},
		{
			"smtps://example.com:465",
		},
		{
			"smtp+tls://example.com:587",
		},
		{ // invalid schema
			"http://example.com:587",
		},
	}
	for i, c := range cases {
		d := newDsn(c.dsn)
		if d.dsnStr != c.dsn {
			t.Errorf(`[Case%d] Result:%v Expected:%v`, i, d.dsnStr, c.dsn)
		}
	}
}

func TestDsnParse(t *testing.T) {
	// Cases
	cases := []struct {
		dsn    string
		scheme string
		host   string
		port   int
		socket string
	}{
		{
			"smtp://example.com:25",
			"smtp",
			"example.com",
			25,
			"example.com:25",
		},
		{
			"smtps://example.com:465",
			"smtps",
			"example.com",
			465,
			"example.com:465",
		},
		{
			"smtp+tls://example.com:587",
			"smtp+tls",
			"example.com",
			587,
			"example.com:587",
		},
		{
			"smtp://127.0.0.1:25",
			"smtp",
			"127.0.0.1",
			25,
			"127.0.0.1:25",
		},
		{
			"smtps://127.0.0.1:465",
			"smtps",
			"127.0.0.1",
			465,
			"127.0.0.1:465",
		},
		{
			"smtp+tls://127.0.0.1:587",
			"smtp+tls",
			"127.0.0.1",
			587,
			"127.0.0.1:587",
		},
		{
			"smtp://[::1]:25",
			"smtp",
			"::1",
			25,
			"[::1]:25",
		},
		{
			"smtps://[::1]:465",
			"smtps",
			"::1",
			465,
			"[::1]:465",
		},
		{
			"smtp+tls://::1:587",
			"smtp+tls",
			"::1",
			587,
			"::1:587",
		},
		{ // upper boundary of a valid TCP port
			"smtp://example.com:65535",
			"smtp",
			"example.com",
			65535,
			"example.com:65535",
		},
	}
	// Test
	for i, c := range cases {
		d := newDsn(c.dsn)
		_ = d.parse()
		if d.scheme != c.scheme {
			t.Errorf(`[Case%d] Scheme Result:%v Expected:%v`, i, d.scheme, c.scheme)
		}
		if d.host != c.host {
			t.Errorf(`[Case%d] Host Result:%v Expected:%v`, i, d.host, c.host)
		}
		if d.port != c.port {
			t.Errorf(`[Case%d] Port Result:%v Expected:%v`, i, d.port, c.port)
		}
		if d.socket != c.socket {
			t.Errorf(`[Case%d] Socket Result:%v Expected:%v`, i, d.socket, c.socket)
		}
	}
}

func TestDsnParseError(t *testing.T) {
	cases := []struct {
		dsn string
		err string
	}{
		{
			"",
			"scheme must be",
		},
		{
			"smtps+tls",
			"scheme must be",
		},
		{
			"smtps://",
			"host must be",
		},
		{
			"smtps://example.com",
			"port must be",
		},
		{
			"smtps://example.com:12345678901234567890",
			"port must be",
		},
		{ // in-range for uint32 but out of the valid TCP port range
			"smtps://example.com:70000",
			"port must be",
		},
	}
	for i, c := range cases {
		d := newDsn(c.dsn)
		err := d.parse()
		if err == nil {
			t.Errorf(`[Case%d] Result: <nil> Expected an error containing:%v`, i, c.err)
			continue
		}
		if !strings.Contains(err.Error(), c.err) {
			t.Errorf(`[Case%d] Result:%v Expected:%v`, i, err, c.err)
		}
	}
}
