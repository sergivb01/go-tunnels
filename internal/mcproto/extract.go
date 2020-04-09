package mcproto

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

const ending = "tunnel.sergitest.dev"

var errNoSrvFound = errors.New("no srv record found")

func ExtractHostPort(serverAddress string) (string, error) {
	h := strings.TrimSuffix(serverAddress, "."+ending)

	_, addrs, err := net.LookupSRV("minecraft", "tcp", h)
	if err != nil {
		if len(addrs) == 0 && strings.Contains(err.Error(), "dnsquery: DNS name does not exist.") {
			return net.JoinHostPort(h, "25565"), nil
		}
		return h, errors.New("unable to find SRV record")
	}

	return net.JoinHostPort(addrs[0].Target, strconv.Itoa(int(addrs[0].Port))), nil
}
