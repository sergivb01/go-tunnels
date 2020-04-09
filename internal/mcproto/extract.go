package mcproto

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	CustomEnding  = os.Getenv("CUSTOM_ENDING")
	errNoSrvFound = errors.New("no srv record found")
)

func ExtractHostPort(serverAddress string) (string, error) {
	h := strings.TrimSuffix(serverAddress, "."+CustomEnding)

	_, addrs, err := net.LookupSRV("minecraft", "tcp", h)
	if err != nil {
		if strings.Contains(err.Error(), "dnsquery: DNS name does not exist.") || strings.Contains(err.Error(), "no such host") {
			return net.JoinHostPort(h, "25565"), nil
		}
		fmt.Println(err)
		return h, errNoSrvFound
	}

	return net.JoinHostPort(addrs[0].Target, strconv.Itoa(int(addrs[0].Port))), nil
}
