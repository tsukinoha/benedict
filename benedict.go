package benedict

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/smtp"
	"strings"
)

const (
	ProdctName = "MaiLetter Mail Client"
	Version    = "0.4.0"
	br         = "\r\n"
	whiteSpace = " \r\n\t\v\b"
	shouldBr   = 78
	mustBr     = 998
)

type (
	AuthInterface interface {
		smtp.Auth
	}
)

func removeBreak(s string) string {
	for _, search := range []string{"\r", "\n"} {
		s = strings.ReplaceAll(s, search, "")
	}
	return s
}

func encodeMime(b []byte, flg bool) []byte {
	needsEnc := false
	for _, v := range b {
		if v > 127 {
			needsEnc = true
			break
		}
	}
	if !needsEnc {
		return b
	}
	if flg {
		// src = bytes.Join([][]byte{[]byte("=?UTF-8?B?"), b, []byte("?=")}, []byte(""))
		dst := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
		base64.StdEncoding.Encode(dst, b)
		buf := []byte("=?UTF-8?B?")
		buf = append(buf, dst...)
		buf = append(buf, []byte("?=")...)
		return buf
	} else {
		dst := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
		base64.StdEncoding.Encode(dst, b)
		return dst
	}
}

func encodeMimeString(s string, flg bool) string {
	needsEnc := false
	for _, v := range s {
		if v > 127 {
			needsEnc = true
			break
		}
	}
	if !needsEnc {
		return s
	}
	if flg {
		return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(s)))
	} else {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
}

func border(length int) string {
	if length < 1 {
		length = 24
	}
	s := []string{
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
		"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	}
	l := len(s)
	sb := strings.Builder{}
	sb.WriteString(strings.Repeat("-", 12))
	for i := 0; i < length; i++ {
		idx := rand.Intn(l)
		sb.WriteString(s[idx])
	}
	return sb.String()
}

func isLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}
