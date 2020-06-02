package mcserver

import (
	"net"
	"strconv"
)

const defaultMCPort = "25565"

func (s MCServer) resolveServerAddress(serverAddress string) (string, *net.TCPAddr, error) {
	var (
		h    = serverAddress[:len(serverAddress)-len(s.cfg.Domain)-1]
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

	return h, addr, nil
}
