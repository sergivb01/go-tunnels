package types

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/sergivb01/mctunnel/internal/protocol/util"
)

var (
	ErrorMismatchedStringLength = errors.New("fewer bytes available than string length")
)

type String string

func (_ String) Decode(r io.Reader) (interface{}, error) {
	size, err := binary.ReadUvarint(util.ByteReader{Reader: r})
	if err != nil {
		return nil, err
	}

	str := make([]byte, int(size))
	if read, err := r.Read(str); err != nil {
		return nil, err
	} else if read != int(size) {
		return nil, ErrorMismatchedStringLength
	}

	return String(str), nil
}

func (s String) Encode(w io.Writer) (int, error) {
	var n int
	var err error

	length := util.Uvarint(uint32(len(s)))

	written, err := w.Write(length)
	n += written
	if err != nil {
		return written, err
	}

	written, err = w.Write([]byte(s))
	n += written
	if err != nil {
		return written, err
	}

	return written, nil
}
