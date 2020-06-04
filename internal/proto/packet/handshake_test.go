package packet

import (
	"bytes"
	"testing"
)

func TestHandshake_Encode(t *testing.T) {
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

func TestHandshake_EncodeForge(t *testing.T) {
	h := &Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "server.com.tunnel.sergivos.dev\x00FML\x00",
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

	if "server.com.tunnel.sergivos.dev" != dH.ServerAddress {
		t.Errorf("Mismatch error, expected %q but got %q", h.ServerAddress, dH.ServerAddress)
		return
	}

	if h.ServerPort != dH.ServerPort {
		t.Errorf("Mismatch error, expected %d but got %d", h.ServerPort, dH.ServerPort)
		return
	}
}

func BenchmarkHandshake_Encode(t *testing.B) {
	h := &Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "rizonmc.net",
		ServerPort:      25565,
		State:           1,
	}

	w := &bytes.Buffer{}
	for i := 0; i < t.N; i++ {
		if err := h.Encode(w); err != nil {
			t.Errorf("Encode() error %v", err)
			return
		}

		if err := h.Decode(w); err != nil {
			t.Errorf("Decode() error %v", err)
			return
		}

		w.Reset()
	}
}
