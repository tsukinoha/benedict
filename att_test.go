package benedict

import (
	"strings"
	"testing"
)

func TestNewAtt(t *testing.T) {
	r := strings.NewReader("file content")
	a, err := NewAtt(r, "test.txt")
	if err != nil {
		t.Fatalf("NewAtt() error = %v", err)
	}
	if a == nil {
		t.Fatal("NewAtt() returned a nil *Att")
	}
	if a.name != "test.txt" {
		t.Errorf("name = %q, want %q", a.name, "test.txt")
	}
	if a.r != r {
		t.Errorf("r = %v, want the reader passed in", a.r)
	}
}

func TestAttName(t *testing.T) {
	cases := []struct {
		name string
	}{
		{"document.pdf"},
		{"日本語ファイル.txt"},
		{""},
	}
	for i, c := range cases {
		a, err := NewAtt(strings.NewReader(""), c.name)
		if err != nil {
			t.Fatalf("[Case%d] NewAtt() error = %v", i, err)
		}
		if got := a.Name(); got != c.name {
			t.Errorf("[Case%d] Name() = %q, want %q", i, got, c.name)
		}
	}
}

func TestAttEncode(t *testing.T) {
	// Att.Encode() is not implemented yet; it always returns an empty
	// string regardless of the attachment content.
	a, err := NewAtt(strings.NewReader("some binary content"), "file.bin")
	if err != nil {
		t.Fatalf("NewAtt() error = %v", err)
	}
	if got := a.Encode(); got != "" {
		t.Errorf("Encode() = %q, want %q (not yet implemented)", got, "")
	}
}
