package mcserver

import (
	"net"
	"os"
)

const defaultMCPort = 25565

var (
	CustomEnding = "." + os.Getenv("CUSTOM_ENDING")
	endLen       = len(CustomEnding) - 1
	// errNoSrvFound = errors.New("no srv record found")
)

func ExtractHostPort(serverAddress string) (string, int) {
	end := len(serverAddress) - endLen
	_, addrs, err := net.LookupSRV("minecraft", "tcp", serverAddress[:end])
	if err != nil || len(addrs) == 0 {
		return serverAddress[:end], defaultMCPort
	}
	return addrs[0].Target, int(addrs[0].Port)
}
