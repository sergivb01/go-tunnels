package mcserver

import (
	"errors"
	"net"
	"os"
)

var (
	CustomEnding  = "." + os.Getenv("CUSTOM_ENDING")
	endLen        = len(CustomEnding)
	errNoSrvFound = errors.New("no srv record found")
)

const defaultMCPort = 25565

func ExtractHostPort(serverAddress string) (string, int) {
	_, addrs, err := net.LookupSRV("minecraft", "tcp", serverAddress[:len(serverAddress)-endLen])
	if err != nil {
		return serverAddress[:len(serverAddress)-endLen], defaultMCPort
	}
	return addrs[0].Target[:len(addrs[0].Target)-1], int(addrs[0].Port)
}
