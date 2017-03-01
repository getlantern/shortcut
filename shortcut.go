// package shortcut constructs a dialer with IPv4 and IPv6 subnets. When the
// address to dial belongs to one of the subnets, it dial via the direct
// dialer, i.e., the shortcut, and dial the proxiedDialer otherwise.

package shortcut

import (
	"context"
	"net"
)

type Dialer struct {
	proxiedDialer func(ctx context.Context, net, addr string) (net.Conn, error)
	directDialer  func(ctx context.Context, net, addr string) (net.Conn, error)
	v4list        List
	v6list        List
}

func NewDialer(
	ipv4Subnets []string,
	ipv6Subnets []string,
	proxiedDialer func(ctx context.Context, net, addr string) (net.Conn, error),
	directDialer func(ctx context.Context, net, addr string) (net.Conn, error),
) *Dialer {
	return &Dialer{
		proxiedDialer: proxiedDialer,
		directDialer:  directDialer,
		v4list:        NewList(ipv4Subnets),
		v6list:        NewList(ipv6Subnets),
	}
}

func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if d.isDomestic(address) {
		return d.directDialer(ctx, network, address)
	}
	return d.proxiedDialer(ctx, network, address)
}

func (d *Dialer) isDomestic(addr string) (hit bool) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			hit = d.v4list.Contains(ip)
			break
		}
		if ip.To16() != nil {
			hit = d.v6list.Contains(ip)
			break
		}
	}
	return
}
