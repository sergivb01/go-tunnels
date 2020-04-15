package mcserver

import (
	"net"
	"os"
)

const defaultMCPort = 25565

var endLen = len(os.Getenv("CUSTOM_ENDING")) + 1

func ResolveServerAddress(serverAddress string) (string, int) {
	_, addrs, err := net.LookupSRV("minecraft", "tcp", serverAddress[:len(serverAddress)-endLen])
	if err != nil || len(addrs) == 0 {
		return serverAddress[:len(serverAddress)-endLen], defaultMCPort
	}
	return addrs[0].Target, int(addrs[0].Port)
}
