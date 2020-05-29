package packet

import (
	"bytes"
	"testing"
)

func TestServerStatus_Encode(t *testing.T) {
	s := &ServerStatus{
		ServerName: "MineBreach",
		Protocol:   47,
		Motd:       "Minebreach Server",
		Favicon:    "",
	}

	w := &bytes.Buffer{}
	if err := s.Encode(w); err != nil {
		t.Errorf("Encode() error %v", err)
		return
	}
}

func BenchmarkServerStatus_Encode(t *testing.B) {
	s := &ServerStatus{
		ServerName: "MineBreach",
		Protocol:   47,
		Motd:       "Minebreach Server",
		Favicon:    "",
	}

	w := &bytes.Buffer{}
	for i := 0; i < t.N; i++ {
		if err := s.Encode(w); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}

		w.Reset()
	}
}
