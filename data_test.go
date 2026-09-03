package benedict

import (
	"bufio"
	"bytes"
	"io"
	"net/mail"
	"reflect"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestNewData(t *testing.T) {
	data := newData()
	if data == nil {
		t.Errorf(`Data struct is empty`)
	}
	if data.headers == nil || len(data.headers) != 0 || data.hdrOrder == nil || len(data.hdrOrder) != 0 {
		t.Errorf(`Data struct headers should be empty (Headers: %d, Orders: %d)`, len(data.headers), len(data.hdrOrder))
	}
	if data.from != nil {
		t.Errorf(`"From" should be empty %v`, data.from)
	}
	if data.returnPath != nil {
		t.Errorf(`"Return-Path" should be empty %v`, data.returnPath)
	}
	if data.replyTo != nil {
		t.Errorf(`"Reply-To" should be empty %v`, data.replyTo)
	}
	if len(data.to) != 0 {
		t.Errorf(`"To" should be empty %v`, data.to)
	}
	if len(data.cc) != 0 {
		t.Errorf(`"Cc" should be empty %v`, data.cc)
	}
	if len(data.bcc) != 0 {
		t.Errorf(`"Bcc" should be empty %v`, data.bcc)
	}
	if data.subject != nil {
		t.Errorf(`"Subject" should be empty %v`, data.subject)
	}
	if data.body != nil {
		t.Errorf(`Body should be empty %v`, data.body)
	}
	if len(data.vars) != 0 {
		t.Errorf(`Variables should be empty %v`, data.vars)
	}
}

func TestDataReset(t *testing.T) {
	data := newData()
	data.headers["xtest"] = header{key: "X-Test", value: "X Test Value"}
	data.hdrOrder = append(data.hdrOrder, "xtest")
	data.from = newAddress("from@example.com", "FromAddress")
	data.to = append(data.to, newAddress("to@example.com", "ToAddress"))
	data.cc = append(data.cc, newAddress("cc@example.com", "CcAddress"))
	data.bcc = append(data.bcc, newAddress("bcc@example.com", "BccAddress"))
	data.subject = template.Must(template.New("Subject").Parse("Test Subject"))
	data.body = template.Must(template.New("Body").Parse("Test Body"))
	data.vars["Test"] = "test value"
	data.reset()
	if data.headers == nil || len(data.headers) != 0 || data.hdrOrder == nil || len(data.hdrOrder) != 0 {
		t.Errorf(`Data struct headers should be empty (Headers: %d, Orders: %d)`, len(data.headers), len(data.hdrOrder))
	}
	if data.from != nil {
		t.Errorf(`"From" should be empty %v`, data.from)
	}
	if data.returnPath != nil {
		t.Errorf(`"Return-Path" should be empty %v`, data.returnPath)
	}
	if data.replyTo != nil {
		t.Errorf(`"Reply-To" should be empty %v`, data.replyTo)
	}
	if len(data.to) != 0 {
		t.Errorf(`"To" should be empty %v`, data.to)
	}
	if len(data.cc) != 0 {
		t.Errorf(`"Cc" should be empty %v`, data.cc)
	}
	if len(data.bcc) != 0 {
		t.Errorf(`"Bcc" should be empty %v`, data.bcc)
	}
	if data.subject != nil {
		t.Errorf(`"Subject" should be empty %v`, data.subject)
	}
	if data.body != nil {
		t.Errorf(`Body should be empty %v`, data.body)
	}
	if len(data.vars) != 0 {
		t.Errorf(`Variables should be empty %v`, data.vars)
	}
}

func TestDataSetHeader(t *testing.T) {
	cases := []struct {
		key    string
		val    string
		exists bool
	}{
		{"Date", "Tue, 5 Mar 2024 21:53:04 +0900", false},
		{"From", "Test From <from@example.com>", false},
		{"To", "Test To <to@example.com>", false},
		{"Cc", "Test Cc <cc@example.com>", false},
		{"Bcc", "bcc@example.com", false},
		{"Subject", "Test Subject", false},
		{"Reply-To", "reply-to@example.com", false},
		{"Return-Path", "return-path@example.com", false},
		{"X-Mailer", "Test MTU 1", true},
		{"Message-ID", "<1234567890ABCDEFGHIJKLMN@example.com>", true},
		{"X-Mailer", "Test MTU 2", true},
	}
	count := 2
	m := newData()
	from := newAddress("from@example.com", "送信者")
	m.setFrom(from)
	for k, v := range cases {
		m.setHeader(v.key, v.val)
		key := strings.ReplaceAll(strings.ToLower(v.key), "-", "")
		_, ok := m.headers[key]
		if ok == v.exists {
			if !v.exists {
				continue
			}
		} else {
			t.Errorf(`[Case%d] "%s" doesn't exist.`, k, v.key)
		}
		if m.headers[key].key != v.key {
			t.Errorf(`[Case%d] Key: %s (%s)`, k, m.headers[key].key, v.key)
		}
		if m.headers[key].value != v.val {
			t.Errorf(`[Case%d] Value: %s (%s)`, k, m.headers[key].value, v.val)
		}
	}
	if len(m.headers) != count {
		t.Errorf("[Case%d] Count: %d (%d)", 999, len(m.headers), count)
	}
}

func TestDataSetFrom(t *testing.T) {
	t.Run("valid address is accepted and seeds ReturnPath/ReplyTo", func(t *testing.T) {
		m := newData()
		from := newAddress("from@example.com", "Sender")
		if err := m.setFrom(from); err != nil {
			t.Fatalf("setFrom() error = %v", err)
		}
		if m.from != from {
			t.Errorf("from = %v, want %v", m.from, from)
		}
		if m.returnPath != from {
			t.Errorf("returnPath = %v, want it defaulted to %v", m.returnPath, from)
		}
		if m.replyTo != from {
			t.Errorf("replyTo = %v, want it defaulted to %v", m.replyTo, from)
		}
	})

	t.Run("existing ReturnPath/ReplyTo are not overwritten", func(t *testing.T) {
		m := newData()
		returnPath := newAddress("return-path@example.com", "")
		replyTo := newAddress("reply-to@example.com", "")
		if err := m.setReturnPath(returnPath); err != nil {
			t.Fatalf("setReturnPath() error = %v", err)
		}
		if err := m.setReplyTo(replyTo); err != nil {
			t.Fatalf("setReplyTo() error = %v", err)
		}
		from := newAddress("from@example.com", "Sender")
		if err := m.setFrom(from); err != nil {
			t.Fatalf("setFrom() error = %v", err)
		}
		if m.returnPath != returnPath {
			t.Errorf("returnPath = %v, want it to stay %v", m.returnPath, returnPath)
		}
		if m.replyTo != replyTo {
			t.Errorf("replyTo = %v, want it to stay %v", m.replyTo, replyTo)
		}
	})

	t.Run("invalid (empty) address is rejected", func(t *testing.T) {
		m := newData()
		if err := m.setFrom(newAddress("", "")); err == nil {
			t.Error("setFrom() error = nil, want an error for an empty address")
		}
		if m.from != nil {
			t.Errorf("from = %v, want nil after a rejected address", m.from)
		}
	})
}

func TestDataSetTo(t *testing.T) {
	cases := []struct {
		addr string
	}{
		{"to+0@example.com"},
		{"to+1@example.com"},
		{"to+2@example.com"},
	}
	m := newData()
	from := newAddress("from@example.com", "送信者")
	m.setFrom(from)
	for k, v := range cases {
		a := newAddress(v.addr, "")
		if err := m.setTo(a); err != nil {
			t.Fatalf("[Case%d] setTo() error = %v", k, err)
		}
		if m.to[k] != a {
			t.Errorf("[Case%d] Address: %v (%v)", k, m.to[k], a)
		}
	}
	if len(m.to) != len(cases) {
		t.Errorf("[Case%d] Count: %d (%d)", 999, len(m.to), len(cases))
	}
	if err := m.setTo(newAddress("", "")); err == nil {
		t.Error(`setTo() with an empty address should return an error`)
	}
	if len(m.to) != len(cases) {
		t.Errorf(`an empty address should not be appended: Count: %d (%d)`, len(m.to), len(cases))
	}
}

func TestDataSetCc(t *testing.T) {
	cases := []struct {
		addr string
	}{
		{"cc+0@example.com"},
		{"cc+1@example.com"},
		{"cc+2@example.com"},
	}
	m := newData()
	from := newAddress("from@example.com", "Sender")
	m.setFrom(from)
	for k, v := range cases {
		a := newAddress(v.addr, "")
		if err := m.setCc(a); err != nil {
			t.Fatalf("[Case%d] setCc() error = %v", k, err)
		}
		if m.cc[k] != a {
			t.Errorf("[Case%d] Address: %v (%v)", k, m.cc[k], a)
		}
	}
	if len(m.cc) != len(cases) {
		t.Errorf("[Case%d] Count: %d (%d)", 999, len(m.cc), len(cases))
	}
	if err := m.setCc(newAddress("", "")); err == nil {
		t.Error(`setCc() with an empty address should return an error`)
	}
	if len(m.cc) != len(cases) {
		t.Errorf(`an empty address should not be appended: Count: %d (%d)`, len(m.cc), len(cases))
	}
}

func TestDataSetBcc(t *testing.T) {
	cases := []struct {
		addr string
	}{
		{
			"bcc+0@example.com",
		},
		{
			"bcc+1@example.com",
		},
		{
			"bcc+2@example.com",
		},
	}
	m := newData()
	from := newAddress("from@example.com", "Sender")
	m.setFrom(from)
	for k, v := range cases {
		a := newAddress(v.addr, "")
		if err := m.setBcc(a); err != nil {
			t.Fatalf("[Case%d] setBcc() error = %v", k, err)
		}
		if m.bcc[k] != a {
			t.Errorf("[Case%d] Address: %v (%v)", k, m.bcc[k], a)
		}
	}
	if len(m.bcc) != len(cases) {
		t.Errorf("[Case%d] Count: %d (%d)", 999, len(m.bcc), len(cases))
	}
	if err := m.setBcc(newAddress("", "")); err == nil {
		t.Error(`setBcc() with an empty address should return an error`)
	}
	if len(m.bcc) != len(cases) {
		t.Errorf(`an empty address should not be appended: Count: %d (%d)`, len(m.bcc), len(cases))
	}
}

func TestDataSetReturnPath(t *testing.T) {
	cases := []struct {
		from *Address
		ret  *Address
	}{
		{
			newAddress("from@example.com", ""),
			newAddress("return-path@example.com", ""),
		},
	}

	for k, v := range cases {
		m := newData()
		if m.returnPath != nil {
			t.Errorf(`[Case%d] InitAddr should be nil before "From:" is set: %v`, k, m.returnPath)
			continue
		}
		if err := m.setFrom(v.from); err != nil {
			t.Fatalf(`[Case%d] setFrom() error: %v`, k, err)
		}
		if m.returnPath.address != v.from.address {
			t.Errorf(`[Case%d] InitAddr: %s (%s)`, k, m.returnPath.address, v.from.address)
			continue
		}
		if err := m.setReturnPath(v.ret); err != nil {
			t.Fatalf(`[Case%d] setReturnPath() error: %v`, k, err)
		}
		if m.returnPath.address != v.ret.address {
			t.Errorf(`[Case%d] InitAddr: %s (%s)`, k, m.returnPath.address, v.ret.address)
		}
		// "Reply-To:" must stay untouched by setReturnPath().
		if m.replyTo.address != v.from.address {
			t.Errorf(`[Case%d] ReplyTo should be unaffected: %s (%s)`, k, m.replyTo.address, v.from.address)
		}
	}
}

func TestDataSetReplyTo(t *testing.T) {
	cases := []struct {
		from  *Address
		reply *Address
	}{
		{
			newAddress("from@example.com", ""),
			newAddress("reply-to@example.com", ""),
		},
	}

	for k, v := range cases {
		m := newData()
		if m.replyTo != nil {
			t.Errorf(`[Case%d] InitAddr should be nil before "From:" is set: %v`, k, m.replyTo)
			continue
		}
		if err := m.setFrom(v.from); err != nil {
			t.Fatalf(`[Case%d] setFrom() error: %v`, k, err)
		}
		if m.replyTo.address != v.from.address {
			t.Errorf(`[Case%d] InitAddr: %s (%s)`, k, m.replyTo.address, v.from.address)
			continue
		}
		m.setReplyTo(v.reply)
		if m.replyTo.address != v.reply.address {
			t.Errorf(`[Case%d] InitAddr: %s (%s)`, k, m.replyTo.address, v.reply.address)
		}
	}
}

func TestDataSetSubject(t *testing.T) {
	cases := []struct {
		subject  string
		vars     map[string]any
		expected string
	}{
		{
			"Subject1",
			nil,
			"Subject1",
		},
		{
			"Dear {{.Name}}",
			map[string]any{"Name": "Mr. Recipient"},
			"Dear Mr. Recipient",
		},
		{ // line breaks in the raw subject must be stripped before parsing
			"Line1\r\nLine2\n",
			nil,
			"Line1Line2",
		},
	}

	m := newData()
	from := newAddress("from@example.com", "送信者")
	m.setFrom(from)
	for k, v := range cases {
		m.setSubject(v.subject)
		typ := reflect.TypeOf(m.subject).String()
		if typ != "*template.Template" {
			t.Errorf(`[Case%d] %v`, k, m)
			continue
		}
		buf := &bytes.Buffer{}
		if err := m.subject.Execute(buf, v.vars); err != nil {
			t.Errorf(`[Case%d] Execute() error: %v`, k, err)
			continue
		}
		if buf.String() != v.expected {
			t.Errorf(`[Case%d] Result: %q Expected: %q`, k, buf.String(), v.expected)
		}
	}
}

func TestDataSetBody(t *testing.T) {
	cases := []struct {
		body     string
		vars     map[string]any
		expected string
	}{
		{
			"Test Body Part1",
			nil,
			"Test Body Part1",
		},
		{ // bare LF must be normalized to CRLF
			"Test\nBody\nPart2\n{{.Name}}",
			map[string]any{"Name": "value"},
			"Test\r\nBody\r\nPart2\r\nvalue",
		},
		{ // CRLF input must remain CRLF (regression test for the
			// strings.Replacer tie-break bug that used to collapse
			// "\r\n" down to a bare "\n")
			"Test\r\nBody\r\nPart2",
			nil,
			"Test\r\nBody\r\nPart2",
		},
		{ // lone CR (old Mac style) must also be normalized to CRLF
			"Test\rBody\rPart2",
			nil,
			"Test\r\nBody\r\nPart2",
		},
	}

	m := newData()
	from := newAddress("from@example.com", "送信者")
	m.setFrom(from)
	for k, v := range cases {
		m.setBody(strings.NewReader(v.body))
		typ := reflect.TypeOf(m.body).String()
		if typ != "*template.Template" {
			t.Errorf(`[Case%d] %v`, k, m)
			continue
		}
		buf := &bytes.Buffer{}
		if err := m.body.Execute(buf, v.vars); err != nil {
			t.Errorf(`[Case%d] Execute() error: %v`, k, err)
			continue
		}
		if buf.String() != v.expected {
			t.Errorf(`[Case%d] Result: %q Expected: %q`, k, buf.String(), v.expected)
		}
	}
}

func TestDataSetValue(t *testing.T) {
	cases := []struct {
		key string
		val any
	}{
		{
			"a",
			"abc",
		},
		{
			"b",
			1,
		},
		{
			"a",
			[]string{"a", "b", "c"},
		},
	}
	expected := 2

	m := newData()
	from := newAddress("from@example.com", "送信者")
	m.setFrom(from)
	for k, v := range cases {
		m.setValue(v.key, v.val)
		typ1 := reflect.TypeOf(m.vars[v.key])
		typ2 := reflect.TypeOf(v.val)
		if typ1 != typ2 {
			t.Errorf("[Case%d] Type: %v (%v)", k, typ1, typ2)
		}
	}
	if len(m.vars) != expected {
		t.Errorf("[Case%d] Count: %d (%d)", 999, len(m.vars), expected)
	}
}

func TestDataString(t *testing.T) {
	cases := []struct {
		headers map[string]header
		to      []*Address
		cc      []*Address
		bcc     []*Address
		subject string
		body    io.Reader
		vars    map[string]any
	}{
		{
			map[string]header{"test-header": {"Test-Header", "Test Header Value"}},
			[]*Address{newAddress("to0@example.com", "受信者To0"), newAddress("to1@example.com", "受信者To1")},
			[]*Address{newAddress("cc0@example.com", "受信者Cc0"), newAddress("cc1@example.com", "受信者Cc1"), newAddress("cc2@example.com", "受信者Cc2")},
			[]*Address{newAddress("bcc0@example.com", "受信者Bcc0")},
			"[{{.ServiceName}}] {{.Name}}様 新商品のお知らせ",
			strings.NewReader("{{.Name}}様\nいつもご利用ありがとうございます。\n{{.ServiceName}}カスタマーサポートでございます。"),
			map[string]any{"ServiceName": "ECサービス", "Name": "ECサービスユーザー"},
		},
	}

	m := newData()
	from := newAddress("from@example.com", "送信者")
	m.setFrom(from)
	for i, c := range cases {
		for _, v := range c.headers {
			m.setHeader(v.key, v.value)
		}
		for _, v := range c.to {
			m.setTo(v)
		}
		for _, v := range c.cc {
			m.setCc(v)
		}
		for _, v := range c.bcc {
			m.setBcc(v)
		}
		for k, v := range c.vars {
			m.setValue(k, v)
		}
		m.setSubject(c.subject)
		m.setBody(c.body)
		result, err := m.Create()
		if err != nil {
			t.Fatalf(`[Case%d] Create() error: %v`, i, err)
		}

		for _, v := range c.to {
			if !strings.Contains(result, v.String()) {
				t.Errorf(`[Case%d] "To:" section does not contain %q`, i, v.String())
			}
		}
		for _, v := range c.cc {
			if !strings.Contains(result, v.String()) {
				t.Errorf(`[Case%d] "Cc:" section does not contain %q`, i, v.String())
			}
		}
		// Bcc recipients must never leak into the generated message headers.
		for _, v := range c.bcc {
			if strings.Contains(result, v.String()) {
				t.Errorf(`[Case%d] "Bcc:" address %q must not appear in the message`, i, v.String())
			}
		}
	}
}

func TestDataCreateErrors(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(d *data)
		wantErr string
	}{
		{
			name:    `"From:" is missing`,
			setup:   func(d *data) {},
			wantErr: `"From:" address is NOT specified.`,
		},
		{
			name: `"To:" is missing`,
			setup: func(d *data) {
				_ = d.setFrom(newAddress("from@example.com", "Sender"))
			},
			wantErr: `"To:" address is NOT specified.`,
		},
		{
			name: "body is missing",
			setup: func(d *data) {
				_ = d.setFrom(newAddress("from@example.com", "Sender"))
				_ = d.setTo(newAddress("to@example.com", "Recipient"))
			},
			wantErr: `mail body is NOT specified.`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := newData()
			c.setup(d)
			result, err := d.Create()
			if err == nil {
				t.Fatalf("Create() returned no error, want %q (result: %q)", c.wantErr, result)
			}
			if err.Error() != c.wantErr {
				t.Errorf("Create() error = %q, want %q", err.Error(), c.wantErr)
			}
			if result != "" {
				t.Errorf("Create() result = %q, want empty string on error", result)
			}
		})
	}
}

func TestDataCreateSuccess(t *testing.T) {
	d := newData()
	from := newAddress("from@example.com", "送信者")
	to0 := newAddress("to0@example.com", "受信者To0")
	to1 := newAddress("to1@example.com", "受信者To1")
	cc0 := newAddress("cc0@example.com", "受信者Cc0")

	if err := d.setFrom(from); err != nil {
		t.Fatalf("setFrom() error = %v", err)
	}
	if err := d.setTo(to0); err != nil {
		t.Fatalf("setTo() error = %v", err)
	}
	if err := d.setTo(to1); err != nil {
		t.Fatalf("setTo() error = %v", err)
	}
	if err := d.setCc(cc0); err != nil {
		t.Fatalf("setCc() error = %v", err)
	}
	if err := d.setHeader("X-Mailer", "Test Mailer"); err != nil {
		t.Fatalf("setHeader() error = %v", err)
	}
	d.setValue("Name", "テスト太郎")
	d.setValue("ServiceName", "ECサービス")
	d.setSubject("[{{.ServiceName}}] {{.Name}}様 新商品のお知らせ")
	d.setBody(strings.NewReader("{{.Name}}様\nいつもご利用ありがとうございます。\n{{.ServiceName}}カスタマーサポートでございます。"))

	result, err := d.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	msg, err := mail.ReadMessage(bufio.NewReader(strings.NewReader(result)))
	if err != nil {
		t.Fatalf("Create() produced an unparseable message: %v\n---\n%s", err, result)
	}

	if got := msg.Header.Get("Content-Type"); got != "text/plain; charset=UTF-8" {
		t.Errorf(`Content-Type = %q, want "text/plain; charset=UTF-8"`, got)
	}
	if got := msg.Header.Get("X-Mailer"); got != "Test Mailer" {
		t.Errorf(`X-Mailer = %q, want "Test Mailer"`, got)
	}
	if _, err := time.Parse(time.RFC1123Z, msg.Header.Get("Date")); err != nil {
		t.Errorf("Date header %q is not a valid RFC1123Z timestamp: %v", msg.Header.Get("Date"), err)
	}
	if got, want := msg.Header.Get("From"), from.String(); got != want {
		t.Errorf("From = %q, want %q", got, want)
	}

	// The subject must be MIME-encoded exactly once after template execution.
	wantSubject := encodeMimeString("[ECサービス] テスト太郎様 新商品のお知らせ", true)
	if got := msg.Header.Get("Subject"); got != wantSubject {
		t.Errorf("Subject = %q, want %q", got, wantSubject)
	}

	for _, addr := range []*Address{to0, to1} {
		if !strings.Contains(result, addr.String()) {
			t.Errorf("To header does not contain %q\n---\n%s", addr.String(), result)
		}
	}
	if !strings.Contains(result, cc0.String()) {
		t.Errorf("Cc header does not contain %q\n---\n%s", cc0.String(), result)
	}

	wantBody := "テスト太郎様\r\nいつもご利用ありがとうございます。\r\nECサービスカスタマーサポートでございます。"
	gotBody, err := io.ReadAll(msg.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if got := string(gotBody); got != wantBody {
		t.Errorf("Body = %q, want %q", got, wantBody)
	}
}

// TestDataCreateMultipleHeaders is a regression test: Create() used to reuse
// a single strings.Builder across the custom-header loop without resetting
// it between iterations, so each header's line stayed in the builder and got
// re-appended to the message for every subsequent header (e.g. with headers
// X-A and X-B, X-A ended up written out twice).
func TestDataCreateMultipleHeaders(t *testing.T) {
	d := newData()
	if err := d.setFrom(newAddress("from@example.com", "Sender")); err != nil {
		t.Fatalf("setFrom() error = %v", err)
	}
	if err := d.setTo(newAddress("to@example.com", "Recipient")); err != nil {
		t.Fatalf("setTo() error = %v", err)
	}
	d.setBody(strings.NewReader("body"))

	headers := []struct{ key, value string }{
		{"X-A", "a"},
		{"X-B", "b"},
		{"X-C", "c"},
	}
	for _, h := range headers {
		if err := d.setHeader(h.key, h.value); err != nil {
			t.Fatalf("setHeader(%q) error = %v", h.key, err)
		}
	}

	result, err := d.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	for _, h := range headers {
		wantLine := h.key + ": " + h.value
		if n := strings.Count(result, wantLine); n != 1 {
			t.Errorf("header line %q appears %d times, want exactly 1\n---\n%s", wantLine, n, result)
		}
	}
}
