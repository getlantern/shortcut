// package shortcut constructs a dialer with IPv4 and IPv6 subnets. When the
// address to dial belongs to one of the subnets, it dial via the direct
// dialer, i.e., the shortcut, and dial the proxiedDialer otherwise.

package shortcut

import (
	"bufio"
	"io"
	"net"
)

type Shortcut struct {
	v4list list
	v6list list
}

// NewFromReader is a helper to create shortcut from readers. The content
// should be in CIDR format, one entry per line.
func NewFromReader(v4 io.Reader, v6 io.Reader) *Shortcut {
	return New(readLines(v4), readLines(v6))
}

// New creates a new shortcut from the subnets.
func New(ipv4Subnets []string, ipv6Subnets []string) *Shortcut {
	return &Shortcut{
		v4list: newList(ipv4Subnets),
		v6list: newList(ipv6Subnets),
	}
}

func readLines(r io.Reader) []string {
	lines := []string{}
	line := ""
	var err error
	br := bufio.NewReader(r)
	for ; err != nil; line, err = br.ReadString('\n') {
		lines = append(lines, line)
	}

	return lines
}

// Allow checks if the address is allowed to use shortcut.
func (s *Shortcut) Allow(addr string) (hit bool) {
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
