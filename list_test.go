package shortcut

import (
	"bufio"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	l := NewList([]string{
		"1.0.1.0/24",
		"1.0.2.0/23",
		"1.0.8.0/21",
		"1.0.32.0/19",
		"1.1.0.0/24",
		"1.1.2.0/23",
		"1.1.4.0/22",
		"1.1.8.0/24",
		"1.1.9.0/24",
		"1.1.10.0/23",
	})
	assert.True(t, l.Contains(net.ParseIP("1.0.1.9")))
	assert.True(t, l.Contains(net.ParseIP("1.0.3.9")))
	assert.False(t, l.Contains(net.ParseIP("1.0.4.9")))
}

func BenchmarkFindWithRadix(b *testing.B) {
	l := newRadixList(readSubnets())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Contains(net.ParseIP("1.0.4.9"))
	}
}

func BenchmarkFindWithSort(b *testing.B) {
	l := newSortList(readSubnets())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Contains(net.ParseIP("1.0.4.9"))
	}
}

func readSubnets() []string {
	f, _ := os.Open("test_list.txt")
	r := bufio.NewScanner(f)
	subnets := []string{}
	for r.Scan() {
		subnets = append(subnets, r.Text())
	}

	return subnets
}
