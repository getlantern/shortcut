package shortcut

import (
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	l := NewSortList([]string{
		"1.1.0.0/24",
		"1.1.2.0/23",
		"1.1.4.0/22",
		"1.1.4.0/24",
		"1.1.10.0/24",
		"1.1.10.0/23",
		"1.1.15.0/24",
		"1.0.1.0/24",
		"1.0.2.0/23",
		"1.0.8.0/21",
		"1.0.32.0/19",
	})
	// smaller than the smallest network
	assert.False(t, l.Contains(net.ParseIP("1.0.0.254").To4()))
	assert.True(t, l.Contains(net.ParseIP("1.0.1.9").To4()))
	assert.True(t, l.Contains(net.ParseIP("1.0.3.9").To4()))
	assert.False(t, l.Contains(net.ParseIP("1.0.4.9").To4()))
	// belong to the larger one of the two overlapped networks, right before
	// another two overlapped networks.
	assert.True(t, l.Contains(net.ParseIP("1.1.6.254").To4()))
	assert.False(t, l.Contains(net.ParseIP("1.1.8.254").To4()))
	assert.True(t, l.Contains(net.ParseIP("1.1.10.254").To4()))
	// belong to the larger one of the two overlapped networks.
	assert.True(t, l.Contains(net.ParseIP("1.1.11.254").To4()))
	// larger than the largest network
	assert.False(t, l.Contains(net.ParseIP("1.1.16.1").To4()))

	l = NewSortList([]string{
		"fe80::1/64",
		"::/64",
		"2001:230:9000::/33",
		"2001:230:8000::/33",
	})
	assert.True(t, l.Contains(net.ParseIP("2001:230:8001::")))
	assert.False(t, l.Contains(net.ParseIP("2001:230:4001::")))
}

func BenchmarkFindWithSort(b *testing.B) {
	f, _ := os.Open("test_list.txt")
	l := NewSortList(readLines(f))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Contains(net.ParseIP("1.0.4.9"))
	}
}
