package packet

import (
	"bytes"
	"testing"
)

func TestLoginDisconnect_Encode(t *testing.T) {
	p := &LoginDisconnect{Reason: "test123"}

	w := &bytes.Buffer{}
	if err := p.Encode(w); err != nil {
		t.Errorf("Encode() error %v", err)
		return
	}
}

func BenchmarkLoginDisconnect_Encode(b *testing.B) {
	p := &LoginStart{Name: "sergivb01"}

	w := &bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		if err := p.Encode(w); err != nil {
			b.Errorf("Encode() error %v", err)
			return
		}

		w.Reset()
	}
}
