package packet

import (
	"bytes"
	"testing"
)

func TestLoginStart_Encode(t *testing.T) {
	l := &LoginStart{Name: "sergivb01"}

	w := &bytes.Buffer{}
	if err := l.Encode(w); err != nil {
		t.Errorf("Encode() error %v", err)
		return
	}

	dP := &LoginStart{}
	if err := dP.Decode(w); err != nil {
		t.Errorf("Decode() error %v", err)
		return
	}

	if l.Name != dP.Name {
		t.Errorf("Mismatch error, expected %q but got %q", l.Name, dP.Name)
		return
	}
}

func BenchmarkLoginStart_Encode(t *testing.B) {
	l := &LoginStart{Name: "sergivb01"}

	w := &bytes.Buffer{}
	for i := 0; i < t.N; i++ {
		if err := l.Encode(w); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}

		if err := l.Decode(w); err != nil {
			t.Errorf("Decode() error %v", err)
			return
		}

		w.Reset()
	}
}
