package benedict

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type (
	dsn struct {
		dsnStr string
		scheme string
		host   string
		port   int
		socket string
	}
)

func newDsn(dsnStr string) *dsn {
	d := new(dsn)
	d.dsnStr = strings.ToLower(strings.TrimSpace(dsnStr))
	d.scheme = ""
	d.host = ""
	d.port = 0
	d.socket = ""
	return d
}

func (d *dsn) parse() error {
	scheme, host, port := "", "", 0
	u, err := url.Parse(d.dsnStr)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "smtp":
		fallthrough
	case "smtps":
		fallthrough
	case "smtp+tls":
		scheme = u.Scheme
	default:
		return fmt.Errorf(`scheme must be one of "smtp", "smtps", or "smtp+tls"`)
	}
	host = u.Hostname()
	if host == "" {
		return fmt.Errorf(`host must be ether hostname or ip address`)
	}
	if u.Port() != "" {
		tmpPort, err := strconv.ParseUint(u.Port(), 10, 32)
		if err != nil || tmpPort > 65535 {
			return fmt.Errorf(`port must be 0 - 65535`)
		}
		port = int(tmpPort)
	} else {
		return fmt.Errorf(`port must be 0 - 65535`)
	}
	d.dsnStr = u.String()
	d.scheme = scheme
	d.host = host
	d.port = port
	d.socket = u.Host
	return nil
}
