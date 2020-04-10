package packet

import "github.com/sergivb01/mctunnel/internal/protocol/types"

type Handshake struct {
	ProtocolVersion types.UVarint
	ServerAddress   types.String
	ServerPort      types.UShort
	NextState       types.UVarint
}

func (h Handshake) ID() int {
	return 0x00
}
