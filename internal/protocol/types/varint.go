package types

import (
	"encoding/binary"
	"io"

	"github.com/sergivb01/mctunnel/internal/protocol/util"
)

type Varint int32

func (_ Varint) Decode(r io.Reader) (interface{}, error) {
	br := util.ByteReader{Reader: r}

	i, err := binary.ReadVarint(br)
	if err != nil {
		return nil, err
	}

	return Varint(int32(i)), nil
}

func (v Varint) Encode(w io.Writer) (int, error) {
	return w.Write(util.Varint(int32(v)))
}
