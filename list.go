package shortcut

import (
	"net"

	"github.com/armon/go-radix"

	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("shortcut")
)

type list interface {
	// Contains checks if the ip belongs to one of the subnet in the list.
	Contains(ip net.IP) bool
}

type radixList struct {
	root *radix.Tree
}

// newList creates a shortcut list from a list of CIDR subnets in "a.b.c.d/24"
// form.
func newList(subnets []string) list {
	return newRadixList(subnets)
}

func newRadixList(subnets []string) list {
	tree := radix.New()
	for _, s := range subnets {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			log.Debugf("Skip %s: %v", s, err)
			continue
		}
		_, _ = tree.Insert(string(n.IP), n.Mask)
	}
	return &radixList{tree}
}

func (l *radixList) Contains(ip net.IP) bool {
	found := false
	l.root.Walk(func(s string, v interface{}) bool {
		ipnet := net.IPNet{net.IP(s), v.(net.IPMask)}
		if ipnet.Contains(ip) {
			found = true
			return true
		}
		return false // continue walk
	})
	return found
}

type sortList struct {
	sorted []*net.IPNet
}
