package benedict

import (
	"encoding/base64"
	"errors"
	"net/smtp"
)

type (
	Login struct {
		username string
		password string
	}
)

func (auth *Login) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !isLocalhost(server.Name) {
		return "", nil, errors.New("unencrypted connection")
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(auth.username)))
	base64.StdEncoding.Encode(dst, []byte(auth.username))
	return "LOGIN", dst, nil
}

func (auth *Login) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(auth.password)))
	base64.StdEncoding.Encode(dst, []byte(auth.password))
	return dst, nil
}
