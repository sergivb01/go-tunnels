package mcserver

import (
	"errors"
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/net/proxy"
)

var errNoProxies = errors.New("no proxies are available")

type roundRobinSwitcher struct {
	proxyURLs []string
	index     uint32
}

func (r *roundRobinSwitcher) GetProxy() (proxy.Dialer, string, error) {
	index := atomic.AddUint32(&r.index, 1) - 1
	proxyURL := r.proxyURLs[index%uint32(len(r.proxyURLs))]

	dialer, err := proxy.SOCKS5("tcp", proxyURL, nil, &net.Dialer{
		Timeout: time.Second * 3,
	})

	return dialer, proxyURL, err
}

// RoundRobinProxySwitcher creates a proxy switcher function which rotates
// ProxyURLs on every request.
// The proxy type is determined by the URL scheme. "http", "https"
// and "socks5" are supported. If the scheme is empty,
// "http" is assumed.
func RoundRobinProxySwitcher(proxies []string) (*roundRobinSwitcher, error) {
	if len(proxies) < 1 {
		return nil, errNoProxies
	}
	return &roundRobinSwitcher{proxies, 0}, nil
}
