package benedict

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
)

var (
	domain   = "example.com"
	mailHost = fmt.Sprintf("mail.%s", domain)
	dsnSmtp  = fmt.Sprintf("smtp://%s", mailHost)
	dsnSmtps = fmt.Sprintf("smtps://%s:465", mailHost)
	dsnTls   = fmt.Sprintf("smtp+tls://%s:25", mailHost)
)

func TestClientNew(t *testing.T) {
	cases := []struct {
		dsn      string
		expected string
	}{
		{
			dsnSmtp,
			"*mailetter.Client",
		},
		{
			dsnSmtps,
			"*mailetter.Client",
		},
		{
			dsnTls,
			"*mailetter.Client",
		},
	}
	for k, v := range cases {
		m := New(v.dsn)
		if reflect.TypeOf(m).String() != v.expected {
			t.Errorf(`[Case%d] %v`, k, reflect.TypeOf(m))
		}
	}
}

func TestMaiLetterLocalName(t *testing.T) {
	cases := []struct {
		call bool
		name string
	}{
		{
			false,
			"localhost",
		},
		{
			true,
			mailHost,
		},
	}

	for k, v := range cases {
		m := New(dsnSmtp)
		if v.call {
			m.LocalName(v.name)
		}
		if m.localName != v.name {
			t.Errorf(`[Case%d] %s (%s)`, k, m.localName, v.name)
		}
	}
}

func TestClientAuthByPlain(t *testing.T) {
	m := New(dsnSmtp)
	m.AuthByPlain("user", "pass")
	if m.auth == nil {
		t.Fatal("auth should not be nil after AuthByPlain()")
	}
	if got := reflect.TypeOf(m.auth).String(); got != "*smtp.plainAuth" {
		t.Errorf("auth type = %s, want *smtp.plainAuth", got)
	}
}

func TestClientAuthByLogin(t *testing.T) {
	m := New(dsnSmtp)
	m.AuthByLogin("user", "pass")
	login, ok := m.auth.(*Login)
	if !ok {
		t.Fatalf("auth type = %T, want *Login", m.auth)
	}
	if login.username != "user" || login.password != "pass" {
		t.Errorf("login = %+v, want username=user password=pass", login)
	}
}

func TestClientAuthByCramMd5(t *testing.T) {
	m := New(dsnSmtp)
	m.AuthByCramMd5("user", "secret")
	if m.auth == nil {
		t.Fatal("auth should not be nil after AuthByCramMd5()")
	}
	if got := reflect.TypeOf(m.auth).String(); got != "*smtp.cramMD5Auth" {
		t.Errorf("auth type = %s, want *smtp.cramMD5Auth", got)
	}
}

func TestClientHeader(t *testing.T) {
	m := New(dsnSmtp)
	if err := m.Header("X-Mailer", "Test Mailer"); err != nil {
		t.Fatalf("Header() error = %v", err)
	}
	if got := m.data.headers["xmailer"].value; got != "Test Mailer" {
		t.Errorf("header value = %q, want %q", got, "Test Mailer")
	}
	if err := m.Header("From", "x@example.com"); err == nil {
		t.Error("Header() with a reserved key should return an error")
	}
}

func TestClientFrom(t *testing.T) {
	m := New(dsnSmtp)
	if err := m.From("from@example.com", "Sender"); err != nil {
		t.Fatalf("From() error = %v", err)
	}
	if m.data.from == nil || m.data.from.address != "from@example.com" || m.data.from.name != "Sender" {
		t.Errorf("from = %+v, want address=from@example.com name=Sender", m.data.from)
	}
	if err := m.From("", ""); err == nil {
		t.Error("From() with an empty address should return an error")
	}
}

func TestClientTo(t *testing.T) {
	m := New(dsnSmtp)
	if err := m.To("to@example.com", "Recipient"); err != nil {
		t.Fatalf("To() error = %v", err)
	}
	if len(m.data.to) != 1 || m.data.to[0].address != "to@example.com" || m.data.to[0].name != "Recipient" {
		t.Errorf("to = %+v, want one address=to@example.com name=Recipient", m.data.to)
	}
	if err := m.To("", ""); err == nil {
		t.Error("To() with an empty address should return an error")
	}
}

func TestClientCc(t *testing.T) {
	m := New(dsnSmtp)
	if err := m.Cc("cc@example.com", "Cc Recipient"); err != nil {
		t.Fatalf("Cc() error = %v", err)
	}
	if len(m.data.cc) != 1 || m.data.cc[0].address != "cc@example.com" || m.data.cc[0].name != "Cc Recipient" {
		t.Errorf("cc = %+v, want one address=cc@example.com name=Cc Recipient", m.data.cc)
	}
	if err := m.Cc("", ""); err == nil {
		t.Error("Cc() with an empty address should return an error")
	}
}

func TestClientBcc(t *testing.T) {
	m := New(dsnSmtp)
	if err := m.Bcc("bcc@example.com", "Bcc Recipient"); err != nil {
		t.Fatalf("Bcc() error = %v", err)
	}
	if len(m.data.bcc) != 1 || m.data.bcc[0].address != "bcc@example.com" || m.data.bcc[0].name != "Bcc Recipient" {
		t.Errorf("bcc = %+v, want one address=bcc@example.com name=Bcc Recipient", m.data.bcc)
	}
	if err := m.Bcc("", ""); err == nil {
		t.Error("Bcc() with an empty address should return an error")
	}
}

func TestClientSubject(t *testing.T) {
	m := New(dsnSmtp)
	m.Subject("Hello {{.Name}}")
	if m.data.subject == nil {
		t.Fatal("subject should not be nil after Subject()")
	}
	buf := &bytes.Buffer{}
	if err := m.data.subject.Execute(buf, map[string]any{"Name": "World"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if buf.String() != "Hello World" {
		t.Errorf("subject = %q, want %q", buf.String(), "Hello World")
	}
}

func TestClientBody(t *testing.T) {
	m := New(dsnSmtp)
	m.Body("Hello {{.Name}}")
	if m.data.body == nil {
		t.Fatal("body should not be nil after Body()")
	}
	buf := &bytes.Buffer{}
	if err := m.data.body.Execute(buf, map[string]any{"Name": "World"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if buf.String() != "Hello World" {
		t.Errorf("body = %q, want %q", buf.String(), "Hello World")
	}
}

func TestClientSet(t *testing.T) {
	m := New(dsnSmtp)
	m.Set("Name", "value")
	if got := m.data.vars["Name"]; got != "value" {
		t.Errorf("vars[Name] = %v, want %v", got, "value")
	}
}

func TestClientSend(t *testing.T) {
	srv := newMockSMTPServer(t, mockSMTPPlain)

	cases := []struct {
		from    *Address
		to      []*Address
		cc      []*Address
		subject string
		body    string
		vals    map[string]string
	}{
		{
			newAddress("from@example.com", "送信者"),
			[]*Address{newAddress("to+1@example.com", "受信者")},
			[]*Address{newAddress("cc@example.com", "Cc宛先")},
			"テスト件名",
			"{{.Name}}",
			map[string]string{"Name": "Mr. Recipient"},
		},
	}
	for k, v := range cases {
		ml := New(fmt.Sprintf("smtp://%s", srv.addr()))
		if err := ml.From(v.from.address, v.from.name); err != nil {
			t.Fatalf("[Case%d] From() error = %v", k, err)
		}
		for _, to := range v.to {
			if err := ml.To(to.address, to.name); err != nil {
				t.Fatalf("[Case%d] To() error = %v", k, err)
			}
		}
		for _, cc := range v.cc {
			if err := ml.Cc(cc.address, cc.name); err != nil {
				t.Fatalf("[Case%d] Cc() error = %v", k, err)
			}
		}
		ml.Subject(v.subject)
		ml.Body(v.body)
		for key, val := range v.vals {
			ml.Set(key, val)
		}

		if err := ml.Send(); err != nil {
			t.Fatalf("[Case%d] Send() error = %v", k, err)
		}

		if got := srv.getMailFrom(); got != "<"+v.from.address+">" {
			t.Errorf("[Case%d] MAIL FROM = %q, want %q", k, got, "<"+v.from.address+">")
		}
		gotRcpt := srv.getRcptTo()
		wantRcpt := []string{}
		for _, a := range append(append([]*Address{}, v.to...), v.cc...) {
			wantRcpt = append(wantRcpt, "<"+a.address+">")
		}
		if !reflect.DeepEqual(gotRcpt, wantRcpt) {
			t.Errorf("[Case%d] RCPT TO = %v, want %v", k, gotRcpt, wantRcpt)
		}
		if data := srv.getData(); !strings.Contains(data, "Mr. Recipient") {
			t.Errorf("[Case%d] DATA does not contain the rendered body: %q", k, data)
		}
	}
}

func TestClientReset(t *testing.T) {
	srv := newMockSMTPServer(t, mockSMTPPlain)
	m := New(fmt.Sprintf("smtp://%s", srv.addr()))
	if err := m.From("from@example.com", "Sender"); err != nil {
		t.Fatalf("From() error = %v", err)
	}
	if err := m.connect(); err != nil {
		t.Fatalf("connect() error = %v", err)
	}
	if err := m.Reset(); err != nil {
		t.Fatalf("Reset() error = %v", err)
	}
	if m.data.from != nil {
		t.Errorf("data should be reset, from = %+v", m.data.from)
	}
}

func TestClientClose(t *testing.T) {
	srv := newMockSMTPServer(t, mockSMTPPlain)
	m := New(fmt.Sprintf("smtp://%s", srv.addr()))
	if err := m.connect(); err != nil {
		t.Fatalf("connect() error = %v", err)
	}
	if err := m.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestClientIsConnect(t *testing.T) {
	srv := newMockSMTPServer(t, mockSMTPPlain)

	t.Run("not yet connected", func(t *testing.T) {
		ml := New(fmt.Sprintf("smtp://%s", srv.addr()))
		if ml.isConnected() {
			t.Error("isConnected() = true, want false before connect()")
		}
	})

	t.Run("connected", func(t *testing.T) {
		ml := New(fmt.Sprintf("smtp://%s", srv.addr()))
		if err := ml.connect(); err != nil {
			t.Fatalf("connect() error = %v", err)
		}
		if !ml.isConnected() {
			t.Error("isConnected() = false, want true after connect()")
		}
		_ = ml.conn.Close()
	})
}

func TestClientConnect(t *testing.T) {
	srv := newMockSMTPServer(t, mockSMTPPlain)

	m := New(fmt.Sprintf("smtp://%s", srv.addr()))
	if err := m.connect(); err != nil {
		t.Fatalf("connect() error = %v", err)
	}
	if !m.isConnected() {
		t.Error("expected isConnected() to be true after connect()")
	}
	// connect() must be a no-op once already connected.
	if err := m.connect(); err != nil {
		t.Errorf("second connect() error = %v", err)
	}
	_ = m.conn.Close()

	t.Run("invalid dsn scheme", func(t *testing.T) {
		m := New("http://example.com:80")
		if err := m.connect(); err == nil {
			t.Error("connect() with an unsupported scheme should return an error")
		}
	})
}

func TestClientConnectBySmtps(t *testing.T) {
	srv := newMockSMTPServer(t, mockSMTPTLS)
	dsnStr := fmt.Sprintf("smtps://%s", srv.addr())
	d := newDsn(dsnStr)
	if err := d.parse(); err != nil {
		t.Fatalf("dsn.parse() error = %v", err)
	}

	ml := New(dsnStr)
	conn, err := ml.connectBySmtps(d)
	if err != nil {
		t.Fatalf("connectBySmtps() error = %v", err)
	}
	defer conn.Close()
	if typ := reflect.TypeOf(conn).String(); typ != "*smtp.Client" {
		t.Errorf("type = %s, want *smtp.Client", typ)
	}
	if err := conn.Hello("localhost"); err != nil {
		t.Errorf("Hello() over the TLS connection failed: %v", err)
	}
}

func TestClientConnectBySmtpsError(t *testing.T) {
	// Nothing is listening here, so the dial itself must fail.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	dsnStr := fmt.Sprintf("smtps://%s", addr)
	d := newDsn(dsnStr)
	if err := d.parse(); err != nil {
		t.Fatalf("dsn.parse() error = %v", err)
	}
	ml := New(dsnStr)
	if _, err := ml.connectBySmtps(d); err == nil {
		t.Error("connectBySmtps() to a closed port should return an error")
	}
}

func TestClientConnectBySmtp(t *testing.T) {
	srv := newMockSMTPServer(t, mockSMTPPlain)
	dsnStr := fmt.Sprintf("smtp://%s", srv.addr())
	d := newDsn(dsnStr)
	if err := d.parse(); err != nil {
		t.Fatalf("dsn.parse() error = %v", err)
	}

	ml := New(dsnStr)
	conn, err := ml.connectBySmtp(d)
	if err != nil {
		t.Fatalf("connectBySmtp() error = %v", err)
	}
	defer conn.Close()
	if typ := reflect.TypeOf(conn).String(); typ != "*smtp.Client" {
		t.Errorf("type = %s, want *smtp.Client", typ)
	}
	if err := conn.Hello("localhost"); err != nil {
		t.Errorf("Hello() failed: %v", err)
	}
}

func TestClientConnectBySmtpError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	dsnStr := fmt.Sprintf("smtp://%s", addr)
	d := newDsn(dsnStr)
	if err := d.parse(); err != nil {
		t.Fatalf("dsn.parse() error = %v", err)
	}
	ml := New(dsnStr)
	if _, err := ml.connectBySmtp(d); err == nil {
		t.Error("connectBySmtp() to a closed port should return an error")
	}
}

func TestClientConnectWithTls(t *testing.T) {
	srv := newMockSMTPServer(t, mockSMTPStartTLS)
	dsnStr := fmt.Sprintf("smtp+tls://%s", srv.addr())
	d := newDsn(dsnStr)
	if err := d.parse(); err != nil {
		t.Fatalf("dsn.parse() error = %v", err)
	}

	ml := New(dsnStr)
	conn, err := ml.connectWithTls(d)
	if err != nil {
		t.Fatalf("connectWithTls() error = %v", err)
	}
	defer conn.Close()
	if typ := reflect.TypeOf(conn).String(); typ != "*smtp.Client" {
		t.Errorf("type = %s, want *smtp.Client", typ)
	}
	if err := conn.Mail("from@example.com"); err != nil {
		t.Errorf("MAIL FROM over the upgraded connection failed: %v", err)
	}
}
