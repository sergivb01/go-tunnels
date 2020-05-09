package packet

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestServerStatus_Encode(t *testing.T) {
	p := &ServerStatus{
		ServerName: "Minebreach",
		Protocol:   47,
		Motd:       "Skere",
		Favicon:    "",
	}
	w := &bytes.Buffer{}
	if err := p.Encode(w); err != nil {
		t.Errorf("Encode() error %v", err)
		return
	}
}

func BenchmarkServerStatus_Encode(t *testing.B) {
	p := &LoginStart{Name: "sergivb01"}

	for i := 0; i < t.N; i++ {
		if err := p.Encode(ioutil.Discard); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}
	}
}
