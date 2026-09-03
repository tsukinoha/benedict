package benedict

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// generateTestCert creates a throw-away self-signed certificate for
// 127.0.0.1, valid for the lifetime of the test process. The client code
// under test always dials with InsecureSkipVerify, so the certificate does
// not need to chain to any trusted root.
func generateTestCert(t *testing.T) tls.Certificate {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create test certificate: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("failed to marshal test key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to build test certificate: %v", err)
	}
	return cert
}

// mockSMTPServer is a minimal SMTP server used to exercise Client's
// connection and Send() logic without touching a real network or a real
// mail server. It supports three modes: plain text, TLS from the first
// byte (smtps), and STARTTLS upgrade (smtp+tls).
type mockSMTPServer struct {
	ln net.Listener

	startTLS bool
	cert     tls.Certificate

	mu       sync.Mutex
	mailFrom string
	rcptTo   []string
	data     string
}

type mockSMTPMode int

const (
	mockSMTPPlain mockSMTPMode = iota
	mockSMTPTLS
	mockSMTPStartTLS
)

func newMockSMTPServer(t *testing.T, mode mockSMTPMode) *mockSMTPServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock SMTP listener: %v", err)
	}

	s := &mockSMTPServer{}
	switch mode {
	case mockSMTPPlain:
		s.ln = ln
	case mockSMTPStartTLS:
		s.startTLS = true
		s.cert = generateTestCert(t)
		s.ln = ln
	case mockSMTPTLS:
		s.cert = generateTestCert(t)
		s.ln = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{s.cert}})
	}

	go s.serve()
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

// addr returns the "host:port" the server is listening on, suitable for
// building a dsn string such as "smtp://" + addr().
func (s *mockSMTPServer) addr() string {
	return s.ln.Addr().String()
}

func (s *mockSMTPServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *mockSMTPServer) handle(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	writeLine := func(line string) {
		w.WriteString(line + "\r\n")
		w.Flush()
	}

	writeLine("220 mock.local ESMTP ready")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		upper := strings.ToUpper(line)

		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			if s.startTLS {
				writeLine("250-mock.local")
				writeLine("250 STARTTLS")
			} else {
				writeLine("250 mock.local")
			}
		case strings.HasPrefix(upper, "STARTTLS"):
			writeLine("220 Ready to start TLS")
			tlsConn := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{s.cert}})
			if err := tlsConn.Handshake(); err != nil {
				return
			}
			conn = tlsConn
			r = bufio.NewReader(conn)
			w = bufio.NewWriter(conn)
		case strings.HasPrefix(upper, "MAIL FROM:"):
			s.mu.Lock()
			s.mailFrom = strings.TrimSpace(line[len("MAIL FROM:"):])
			s.mu.Unlock()
			writeLine("250 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			s.mu.Lock()
			s.rcptTo = append(s.rcptTo, strings.TrimSpace(line[len("RCPT TO:"):]))
			s.mu.Unlock()
			writeLine("250 OK")
		case strings.HasPrefix(upper, "DATA"):
			writeLine("354 End data with <CR><LF>.<CR><LF>")
			var body strings.Builder
			for {
				dl, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if dl == ".\r\n" || dl == ".\n" {
					break
				}
				body.WriteString(dl)
			}
			s.mu.Lock()
			s.data = body.String()
			s.mu.Unlock()
			writeLine("250 OK: queued")
		case strings.HasPrefix(upper, "RSET"):
			writeLine("250 OK")
		case strings.HasPrefix(upper, "NOOP"):
			writeLine("250 OK")
		case strings.HasPrefix(upper, "QUIT"):
			writeLine("221 Bye")
			return
		default:
			writeLine(fmt.Sprintf("500 unrecognized command: %s", line))
		}
	}
}

func (s *mockSMTPServer) getMailFrom() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mailFrom
}

func (s *mockSMTPServer) getRcptTo() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.rcptTo))
	copy(out, s.rcptTo)
	return out
}

func (s *mockSMTPServer) getData() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}
