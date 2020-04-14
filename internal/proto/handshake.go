package proto

import (
	"io"
)

type Handshake struct {
	ProtocolVersion uint64
	Address         string
	Port            uint16
	Status          uint64
}

func (h *Handshake) Write(remote io.Writer, address string) error {
	w, err := packetWriter(0x00)
	if err != nil {
		return err
	}

	if err := writeVarInt(w, h.ProtocolVersion); err != nil {
		return err
	}

	if err := writeString(w, address); err != nil {
		return err
	}

	if err := writeShort(w, h.Port); err != nil {
		return err
	}

	if err := writeVarInt(w, h.Status); err != nil {
		return err
	}

	if err := writeVarInt(remote, uint64(w.Len())); err != nil {
		return err
	}

	if _, err := remote.Write(w.Bytes()); err != nil {
		return err
	}

	return nil
}

func (h *Handshake) ReverseWrite(remote io.Writer, address string) error {
	w, err := packetWriter(0x00)
	if err != nil {
		return err
	}

	if err := writeVarInt(w, h.ProtocolVersion); err != nil {
		return err
	}

	if err := writeString(w, address); err != nil {
		return err
	}

	if err := writeShort(w, h.Port); err != nil {
		return err
	}

	if err := writeVarInt(w, h.Status); err != nil {
		return err
	}

	if err := writeVarInt(remote, uint64(w.Len())); err != nil {
		return err
	}

	if _, err := w.WriteTo(remote); err != nil {
		return err
	}

	return nil
}

func ReadHandshake(reader BytesReader) (*Handshake, error) {
	version, err := readVarInt(reader)
	if err != nil {
		return nil, err
	}

	clientAddress, err := readString(reader)
	if err != nil {
		return nil, err
	}

	clientPort, err := readShort(reader)
	if err != nil {
		return nil, err
	}

	request, err := readVarInt(reader)

	if err != nil {
		return nil, err
	}

	return &Handshake{
		ProtocolVersion: version,
		Address:         clientAddress,
		Port:            clientPort,
		Status:          request,
	}, nil
}
