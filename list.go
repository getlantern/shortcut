package shortcut

import (
	"bytes"
	"net"
	"sort"

	"github.com/armon/go-radix"

	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("shortcut")
)

type List interface {
	// Contains checks if the ip belongs to one of the subnet in the list.
	Contains(ip net.IP) bool
}

type radixList struct {
	root *radix.Tree
}

// NewList creates a shortcut list from a list of CIDR subnets in "a.b.c.d/24"
// form.
func NewList(subnets []string) List {
	return newRadixList(subnets)
}

func newRadixList(subnets []string) List {
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

// It's used as a benchmark baseline for radix implementation.
func newSortList(subnets []string) List {
	nets := make([]*net.IPNet, 0, len(subnets))
	for _, s := range subnets {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			log.Debugf("Skip %s: %v", s, err)
			continue
		}

		nets = append(nets, n)
	}
	sort.Slice(nets, func(i, j int) bool {
		r := bytes.Compare(nets[i].IP, nets[j].IP)
		switch r {
		case -1:
			return true
		case 1:
			return false
		default:
			return bytes.Compare(nets[i].Mask, nets[j].Mask) > 0
		}
	})
	return &sortList{nets}
}

func (l *sortList) Contains(ip net.IP) bool {
	index := sort.Search(len(l.sorted), func(i int) bool {
		return l.sorted[i].Contains(ip)
	})
	return index != len(l.sorted)
}
