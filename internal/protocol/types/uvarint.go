package types

import (
	"encoding/binary"
	"io"

	"github.com/sergivb01/mctunnel/internal/protocol/util"
)

type UVarint uint32

func (_ UVarint) Decode(r io.Reader) (interface{}, error) {
	br := util.ByteReader{Reader: r}

	i, err := binary.ReadUvarint(br)
	if err != nil {
		return nil, err
	}

	return UVarint(uint32(i)), nil
}

func (u UVarint) Encode(w io.Writer) (int, error) {
	return w.Write(util.Uvarint(uint32(u)))
}
