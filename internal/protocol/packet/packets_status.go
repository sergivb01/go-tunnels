package packet

import (
	"github.com/sergivb01/mctunnel/internal/protocol/chat"
	"github.com/sergivb01/mctunnel/internal/protocol/types"
)

type StatusRequest struct{}

func (r *StatusRequest) ID() int { return 0x00 }

type StatusResponse struct {
	Status struct {
		Version struct {
			Name     string `json:"name"`
			Protocol int    `json:"protocol"`
		} `json:"version"`

		Players struct {
			Max    int `json:"max"`
			Online int `json:"online"`
		} `json:"players"`

		Description chat.TextComponent `json:"description"`
	}
}

func (r StatusResponse) ID() int { return 0x00 }

type StatusPing struct {
	Payload types.Long
}

func (p StatusPing) ID() int { return 0x01 }

type StatusPong struct {
	Payload types.Long
}

func (p StatusPong) ID() int { return 0x01 }
