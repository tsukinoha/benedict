package benedict

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"
)

type (
	header struct {
		key   string
		value string
	}
	data struct {
		headers    map[string]header
		hdrOrder   []string
		from       *Address
		returnPath *Address
		replyTo    *Address
		to         []*Address
		cc         []*Address
		bcc        []*Address
		subject    *template.Template
		body       *template.Template
		vars       map[string]any
	}
)

func newData() *data {
	d := new(data)
	d.reset()
	return d
}

func (d *data) reset() {
	d.headers = map[string]header{}
	d.hdrOrder = []string{}
	d.from = nil
	d.to = []*Address{}
	d.cc = []*Address{}
	d.bcc = []*Address{}
	d.subject = nil
	d.body = nil
	d.vars = map[string]any{}
}

func (d *data) setHeader(key, value string) error {
	key = removeBreak(key)
	key = strings.ReplaceAll(key, ":", "")
	value = removeBreak(value)
	excepts := map[string]bool{
		"contenttype": true,
		"date":        true,
		"from":        true,
		"replyto":     true,
		"returnpath":  true,
		"to":          true,
		"cc":          true,
		"bcc":         true,
		"subject":     true,
	}
	lowerKey := strings.ToLower(key)
	lowerKey = strings.ReplaceAll(lowerKey, "-", "")
	if _, ok := excepts[lowerKey]; ok {
		return fmt.Errorf(`header key "%s" is reserved`, key)
	}

	d.headers[lowerKey] = header{key: key, value: value}
	exists := false
	for _, v := range d.hdrOrder {
		if v == key {
			exists = true
			break
		}
	}
	if !exists {
		d.hdrOrder = append(d.hdrOrder, key)
	}
	return nil
}

func (d *data) setFrom(addr *Address) error {
	if addr.address == "" {
		return fmt.Errorf(`the addess for "From:" is invalid (%s)`, addr.address)
	}
	d.from = addr
	if d.returnPath == nil {
		d.returnPath = addr
	}
	if d.replyTo == nil {
		d.replyTo = addr
	}
	return nil
}

func (d *data) setTo(addr *Address) error {
	if addr.address == "" {
		return fmt.Errorf(`the addess for "To:" is invalid (%s)`, addr.address)
	}
	d.to = append(d.to, addr)
	return nil
}

func (d *data) setCc(addr *Address) error {
	if addr.address == "" {
		return fmt.Errorf(`the addess for "Cc:" is invalid (%s)`, addr.address)
	}
	d.cc = append(d.cc, addr)
	return nil
}

func (d *data) setBcc(addr *Address) error {
	if addr.address == "" {
		return fmt.Errorf(`the addess for "Bcc:" is invalid (%s)`, addr.address)
	}
	d.bcc = append(d.bcc, addr)
	return nil
}

func (d *data) setReturnPath(addr *Address) error {
	if addr.address == "" {
		return fmt.Errorf(`the addess for "Return-Path:" is invalid (%s)`, addr.address)
	}
	d.returnPath = addr
	return nil
}

func (d *data) setReplyTo(addr *Address) error {
	if addr.address == "" {
		return fmt.Errorf(`the addess for "Reply-To:" is invalid (%s)`, addr.address)
	}
	d.replyTo = addr
	return nil
}

func (d *data) setSubject(subject string) {
	for old, new := range map[string]string{"\r": "", "\n": ""} {
		subject = strings.ReplaceAll(subject, old, new)
	}
	d.subject = template.Must(template.New("Subject").Parse(subject))
}

func (d *data) setBody(r io.Reader) {
	body, _ := io.ReadAll(r)
	bodyText := string(body)
	// Normalize every line-ending style to CRLF. This has to happen as three
	// sequential passes: a single strings.Replacer with all three rules would
	// match "\r\n" and "\n" at the same position and (per its documented
	// argument-order tie-break) always prefer the "\r\n"->"\n" rule, which
	// collapses CRLF input down to bare LF instead of normalizing it to CRLF.
	bodyText = strings.ReplaceAll(bodyText, "\r\n", "\n")
	bodyText = strings.ReplaceAll(bodyText, "\r", "\n")
	bodyText = strings.ReplaceAll(bodyText, "\n", "\r\n")
	d.body = template.Must(template.New("Body").Parse(bodyText))
}

func (d *data) setValue(key string, val any) {
	d.vars[key] = val
}

func (d *data) Create() (string, error) {
	if d.from == nil {
		return "", errors.New(`"From:" address is NOT specified.`)
	} else if len(d.to) == 0 {
		return "", errors.New(`"To:" address is NOT specified.`)
	} else if d.body == nil {
		return "", errors.New(`mail body is NOT specified.`)
	}

	sb := strings.Builder{}
	line := strings.Builder{}
	// buf := bytes.NewBuffer(make([]byte, 10240))

	// Headers
	for _, v := range d.headers {
		line.Reset()
		line.WriteString(fmt.Sprintf("%s: %s\r\n", v.key, encodeMimeString(v.value, true)))
		sb.WriteString(line.String())
	}

	// Content-Type
	sb.WriteString("Content-Type: text/plain; charset=UTF-8" + br)
	// Date
	sb.WriteString(fmt.Sprintf("Date: %s%s", time.Now().Format(time.RFC1123Z), br))
	// From
	sb.WriteString(fmt.Sprintf("From: %s%s", encodeMimeString(d.from.String(), true), br))
	// To, Cc
	rcpts := map[string][]*Address{"To": d.to, "Cc": d.cc}
	for label, addrs := range rcpts {
		if len(addrs) == 0 {
			continue
		}
		line.Reset()
		line.WriteString(label)
		line.WriteString(":")
		for k, v := range addrs {
			if k == 0 {
				line.WriteString(" ")
			} else {
				line.WriteString("    ")
			}
			line.WriteString(v.String())
			if len(addrs) > k+1 {
				line.WriteString(",")
			}
			line.WriteString("\r\n")
		}
		sb.WriteString(line.String())
	}

	// Subject
	if d.subject == nil {
		d.setSubject("")
	}
	buf := bytes.NewBuffer([]byte{})
	d.subject.Execute(buf, d.vars)
	line.Reset()
	line.WriteString("Subject: ")
	line.WriteString(encodeMimeString(buf.String(), true))
	line.WriteString("\r\n")
	sb.WriteString(line.String())

	// Body
	buf.Reset()
	if d.body == nil {
		d.setBody(strings.NewReader(""))
	}
	d.body.Execute(buf, d.vars)
	line.Reset()
	line.WriteString("\r\n")
	line.WriteString(buf.String())
	sb.WriteString(line.String())

	return sb.String(), nil
}
