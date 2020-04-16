package mcserver

import (
	"net"
	"os"
	"strconv"

	"github.com/patrickmn/go-cache"
)

const defaultMCPort = "25565"

var endLen = len(os.Getenv("CUSTOM_ENDING")) + 1

type connDetails struct {
	host string
	port string
}

func (s *MCServer) ResolveServerAddress(serverAddress string) (string, string) {
	if details, found := s.c.Get(serverAddress); found {
		return details.(*connDetails).host, details.(*connDetails).port
	}

	var (
		h    = serverAddress[:len(serverAddress)-endLen]
		port = defaultMCPort
	)

	_, addrs, err := net.LookupSRV("minecraft", "tcp", h)
	if err == nil && len(addrs) != 0 {
		h, port = addrs[0].Target, strconv.Itoa(int(addrs[0].Port))
	}

	s.c.Set(serverAddress, &connDetails{
		host: h,
		port: port,
	}, cache.DefaultExpiration)

	return h, port
}
