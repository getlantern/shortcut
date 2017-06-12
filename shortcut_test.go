package shortcut

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllow(t *testing.T) {
	s := NewFromReader(
		strings.NewReader("127.0.0.0/24\n8.8.0.0/16\n"),
		strings.NewReader("fe80::1/64\n::/64\n2001:4860:4860::8800/120\n"),
	)

	allow := func(addr string) bool {
		hit, _ := s.Allow(addr)
		return hit
	}
	assert.True(t, allow("127.0.0.1:8888"))
	assert.True(t, allow("localhost:8888"))
	assert.True(t, allow("localhost"))
	assert.True(t, allow("google-public-dns-a.google.com"))
	assert.True(t, allow("google-public-dns-b.google.com"))
	assert.False(t, allow("1.2.4.5:8888"))
	assert.False(t, allow("not-exist.com"))
}
