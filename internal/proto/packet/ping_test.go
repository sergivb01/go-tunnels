package packet

import (
	"bytes"
	"testing"
)

func TestPing_Encode(t *testing.T) {
	p1 := &Ping{Payload: 1234567890123456789}

	w := &bytes.Buffer{}
	if err := p1.Encode(w); err != nil {
		t.Errorf("Encode() error %v", err)
		return
	}

	p2 := &Ping{}
	if err := p2.Decode(w); err != nil {
		t.Errorf("Decode() error %v", err)
		return
	}

	if p1.Payload != p2.Payload {
		t.Errorf("Mismatch error Payload, expected %d but got %d", p1.Payload, p2.Payload)
		return
	}
}

func BenchmarkPing_Encode(t *testing.B) {
	p := &Ping{Payload: 1234567890123456789}

	w := &bytes.Buffer{}
	for i := 0; i < t.N; i++ {
		if err := p.Encode(w); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}

		if err := p.Decode(w); err != nil {
			t.Errorf("Decode() error %v", err)
			return
		}

		w.Reset()
	}
}
