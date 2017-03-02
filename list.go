package shortcut

import (
	"bytes"
	"fmt"
	"net"

	"github.com/petar/GoLLRB/llrb"

	"github.com/getlantern/golog"
)

var (
	log = golog.LoggerFor("shortcut")
)

type ipList struct {
	root *llrb.LLRB
}

type entry struct {
	ipnet *net.IPNet
}

func (a *entry) Less(b llrb.Item) bool {
	ipa := a.ipnet.IP
	ipb := b.(*entry).ipnet.IP
	return bytes.Compare(ipa, ipb) < 0
}

func (a *entry) String() string {
	return a.ipnet.String()
}

// newipList creates a shortcut list from a list of CIDR subnets in "a.b.c.d/24"
// form, inspired by discussion at
// http://stackoverflow.com/questions/13875486/how-to-sort-ip-addresses-in-a-trie-table
func newIPList(subnets []string) (*ipList, error) {
	tree := llrb.New()
	for _, s := range subnets {
		_, ipnet, err := net.ParseCIDR(s)
		if err != nil {
			log.Debugf("Skip %s: %v", s, err)
			continue
		}
		ipnet.IP = normalize(ipnet.IP)
		tree.InsertNoReplace(&entry{ipnet})
	}

	// Sanity check for overlap
	detectedOverlap := false
	var overlapError error
	tree.DescendLessOrEqual(&entry{
		&net.IPNet{
			IP: net.IP([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}),
		},
	}, func(i llrb.Item) bool {
		e := i.(*entry)
		ip := e.ipnet.IP
		first := true
		tree.DescendLessOrEqual(e, func(i llrb.Item) bool {
			if first {
				// Ignore the first, since it will same as pivot
				first = false
				return true
			}
			if i.(*entry).ipnet.Contains(ip) {
				detectedOverlap = true
				overlapError = fmt.Errorf("Overlapping ip ranges detected, %v contains %v", i, ip)
				return false
			}
			return true
		})
		return !detectedOverlap
	})

	if detectedOverlap {
		return nil, overlapError
	}

	return &ipList{tree}, nil
}

// Contains checks if the ip belongs to one of the subnet in the list.
func (l *ipList) Contains(ip net.IP) bool {
	ip = normalize(ip)
	e := &entry{
		&net.IPNet{
			IP: ip,
		},
	}
	found := false
	l.root.DescendLessOrEqual(e, func(i llrb.Item) bool {
		found = i.(*entry).ipnet.Contains(ip)
		return false
	})
	return found
}

func normalize(ip net.IP) net.IP {
	return ip.To16()
}
