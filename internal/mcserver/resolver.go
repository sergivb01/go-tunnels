package mcserver

import (
	"net"
	"os"
	"strconv"
)

const defaultMCPort = "25565"

var endLen = len(os.Getenv("CUSTOM_ENDING")) + 1

type connDetails struct {
	host string
	addr *net.TCPAddr
}

func (s *MCServer) ResolveServerAddress(serverAddress string) (string, *net.TCPAddr, error) {
	if details, found := s.c.Get(serverAddress); found {
		return details.(*connDetails).host, details.(*connDetails).addr, nil
	}

	var (
		h    = serverAddress[:len(serverAddress)-endLen]
		port = defaultMCPort
	)

	_, addrs, err := net.LookupSRV("minecraft", "tcp", h)
	if err == nil && len(addrs) != 0 {
		h, port = addrs[0].Target, strconv.Itoa(int(addrs[0].Port))
	}

	addr, err := net.ResolveTCPAddr("tcp", h+":"+port)
	if err != nil {
		return "", nil, err
	}

	s.c.SetDefault(serverAddress, &connDetails{
		host: h,
		addr: addr,
	})

	return h, addr, nil
}
