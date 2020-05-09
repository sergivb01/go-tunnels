package packet

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestHandshake_EncodeDecode(t *testing.T) {
	h := &Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "rizonmc.net",
		ServerPort:      25565,
		State:           2,
	}
	w := &bytes.Buffer{}
	if err := h.Encode(w); err != nil {
		t.Errorf("Encode() error %v", err)
		return
	}

	dH := &Handshake{}
	if err := dH.Decode(w); err != nil {
		t.Errorf("Decode() error %v", err)
		return
	}

	if h.ProtocolVersion != dH.ProtocolVersion {
		t.Errorf("Mismatch error, expected %d but got %d", h.ProtocolVersion, dH.ProtocolVersion)
		return
	}

	if h.State != dH.State {
		t.Errorf("Mismatch error, expected %d but got %d", h.State, dH.State)
		return
	}

	if h.ServerAddress != dH.ServerAddress {
		t.Errorf("Mismatch error, expected %q but got %q", h.ServerAddress, dH.ServerAddress)
		return
	}

	if h.ServerPort != dH.ServerPort {
		t.Errorf("Mismatch error, expected %d but got %d", h.ServerPort, dH.ServerPort)
		return
	}
}

func BenchmarkHandshake_EncodeDecode(t *testing.B) {
	h := &Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "rizonmc.net",
		ServerPort:      25565,
		State:           1,
	}

	for i := 0; i < t.N; i++ {
		w := &bytes.Buffer{}
		if err := h.Encode(w); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}

		dH := &Handshake{}
		if err := dH.Decode(w); err != nil {
			t.Errorf("Decode() error %v", err)
			return
		}
	}
}

func BenchmarkHandshake_Encode(t *testing.B) {
	h := &Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "rizonmc.net",
		ServerPort:      25565,
		State:           1,
	}

	for i := 0; i < t.N; i++ {
		if err := h.Encode(ioutil.Discard); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}
	}
}
