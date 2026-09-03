package benedict

import (
	"encoding/base64"
	"net/smtp"
	"testing"
)

func TestLoginStart(t *testing.T) {
	cases := []struct {
		name       string
		serverName string
		tls        bool
		wantErr    bool
	}{
		{
			name:       "plain connection to remote host is rejected",
			serverName: "mail.example.com",
			tls:        false,
			wantErr:    true,
		},
		{
			name:       "TLS connection is allowed",
			serverName: "mail.example.com",
			tls:        true,
			wantErr:    false,
		},
		{
			name:       "plain connection to localhost is allowed",
			serverName: "localhost",
			tls:        false,
			wantErr:    false,
		},
		{
			name:       "plain connection to 127.0.0.1 is allowed",
			serverName: "127.0.0.1",
			tls:        false,
			wantErr:    false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			auth := &Login{username: "user", password: "pass"}
			proto, toServer, err := auth.Start(&smtp.ServerInfo{Name: c.serverName, TLS: c.tls})
			if c.wantErr {
				if err == nil {
					t.Fatalf("Start() error = nil, want an error")
				}
				if proto != "" || toServer != nil {
					t.Errorf("Start() on error = (%q, %v), want (\"\", nil)", proto, toServer)
				}
				return
			}
			if err != nil {
				t.Fatalf("Start() error = %v", err)
			}
			if proto != "LOGIN" {
				t.Errorf("Start() proto = %q, want %q", proto, "LOGIN")
			}
			wantToServer := base64.StdEncoding.EncodeToString([]byte("user"))
			if string(toServer) != wantToServer {
				t.Errorf("Start() toServer = %q, want %q", toServer, wantToServer)
			}
		})
	}
}

func TestLoginNext(t *testing.T) {
	auth := &Login{username: "user", password: "pass"}

	t.Run("more input requested returns the encoded password", func(t *testing.T) {
		toServer, err := auth.Next([]byte("Password:"), true)
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		want := base64.StdEncoding.EncodeToString([]byte("pass"))
		if string(toServer) != want {
			t.Errorf("Next() = %q, want %q", toServer, want)
		}
	})

	t.Run("no more input requested returns nil", func(t *testing.T) {
		toServer, err := auth.Next(nil, false)
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if toServer != nil {
			t.Errorf("Next() = %q, want nil once the exchange is complete", toServer)
		}
	})
}
