package mcserver

import (
	"errors"
	"net"
	"os"
	"strings"
)

var (
	CustomEnding  = os.Getenv("CUSTOM_ENDING")
	errNoSrvFound = errors.New("no srv record found")
)

func ExtractHostPort(serverAddress string) (string, int, error) {
	h := strings.TrimSuffix(serverAddress, "."+CustomEnding)
	_, addrs, err := net.LookupSRV("minecraft", "tcp", h)
	if err != nil {
		if strings.Contains(err.Error(), "dnsquery: DNS name does not exist.") || strings.Contains(err.Error(), "no such host") {
			return h, 25565, nil
		}
		return h, 25565, errNoSrvFound
	}
	return addrs[0].Target[:len(addrs[0].Target)-1], int(addrs[0].Port), nil
}
