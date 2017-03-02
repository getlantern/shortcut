// package shortcut constructs a dialer with IPv4 and IPv6 subnets. When the
// address to dial belongs to one of the subnets, it dial via the direct
// dialer, i.e., the shortcut, and dial the proxiedDialer otherwise.

package shortcut

import (
	"bufio"
	"io"
	"net"
)

type Shortcut interface {
	// Allow checks if the address is allowed to use shortcut.
	Allow(addr string) bool
}

type shortcut struct {
	v4list *ipList
	v6list *ipList
}

// NewFromReader is a helper to create shortcut from readers. The content
// should be in CIDR format, one entry per line.
func NewFromReader(v4 io.Reader, v6 io.Reader) (Shortcut, error) {
	return New(readLines(v4), readLines(v6))
}

func readLines(r io.Reader) []string {
	lines := []string{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

// New creates a new shortcut from the subnets.
func New(ipv4Subnets []string, ipv6Subnets []string) (Shortcut, error) {
	log.Debugf("Creating shortcut with %d ipv4 subnets and %d ipv6 subnets",
		len(ipv4Subnets),
		len(ipv6Subnets),
	)

	v4list, err := newIPList(ipv4Subnets)
	if err != nil {
		return nil, err
	}

	v6list, err := newIPList(ipv6Subnets)
	if err != nil {
		return nil, err
	}

	return &shortcut{
		v4list: v4list,
		v6list: v6list,
	}, nil
}

func (s *shortcut) Allow(addr string) (hit bool) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			hit = s.v4list.Contains(ip)
			break
		}
		if ip.To16() != nil {
			hit = s.v6list.Contains(ip)
			break
		}
	}
	return
}
