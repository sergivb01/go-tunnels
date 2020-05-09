package packet

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestLoginStart_EncodeDecode(t *testing.T) {
	p := &LoginStart{Name: "sergivb01"}
	w := &bytes.Buffer{}
	if err := p.Encode(w); err != nil {
		t.Errorf("Encode() error %v", err)
		return
	}

	dP := &LoginStart{}
	if err := dP.Decode(w); err != nil {
		t.Errorf("Decode() error %v", err)
		return
	}

	if p.Name != dP.Name {
		t.Errorf("Mismatch error, expected %q but got %q", p.Name, dP.Name)
		return
	}
}

func BenchmarkLoginStart_EncodeDecode(t *testing.B) {
	p := &LoginStart{Name: "sergivb01"}

	for i := 0; i < t.N; i++ {
		w := &bytes.Buffer{}
		if err := p.Encode(w); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}

		dH := &LoginStart{}
		if err := dH.Decode(w); err != nil {
			t.Errorf("Decode() error %v", err)
			return
		}
	}
}

func BenchmarkLoginStart_Encode(t *testing.B) {
	p := &LoginStart{Name: "sergivb01"}

	for i := 0; i < t.N; i++ {
		if err := p.Encode(ioutil.Discard); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}
	}
}
