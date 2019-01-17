package shortcut

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAllow(t *testing.T) {
	s := NewFromReader(
		strings.NewReader("127.0.0.0/24\n8.8.0.0/16\n"),
		strings.NewReader("fe80::1/64\n::/64\n2001:4860:4860::8800/120\n"),
	)

	allow := func(addr string) bool {
		hit, _ := s.Allow(context.Background(), addr)
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

func TestContext(t *testing.T) {
	s := NewFromReader(
		strings.NewReader("127.0.0.0/24\n8.8.0.0/16\n"),
		strings.NewReader("fe80::1/64\n::/64\n2001:4860:4860::8800/120\n"),
	)
	s.(*shortcut).resolver.PreferGo = true // to use the supplied Dial() function
	s.(*shortcut).resolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		time.Sleep(100 * time.Millisecond)
		var d net.Dialer
		return d.DialContext(ctx, network, address)
	}
	ctx := context.Background()
	hit, _ := s.Allow(ctx, "google-public-dns-a.google.com:8888")
	assert.True(t, hit, "host should be allowed when IP is in the list")
	ctx, _ = context.WithTimeout(ctx, 10*time.Millisecond)
	hit, _ = s.Allow(ctx, "google-public-dns-a.google.com:8888")
	assert.False(t, hit, "host should be disallowed if context exceeded performing DNS lookup")
}

func TestSystemResolverTiming(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Linux doesn't have DNS cache by default")
	}
	resolver := net.DefaultResolver
	rand.Seed(time.Now().UnixNano())
	host := fmt.Sprintf("host%d.com", rand.Int31())
	fmt.Printf("Looking up %s\n", host)
	start := time.Now()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	resolver.LookupIPAddr(ctx, host)
	lap1 := time.Now()
	ctx, _ = context.WithTimeout(context.Background(), time.Second)
	resolver.LookupIPAddr(ctx, host)
	lap2 := time.Now()
	// `tcpdump udp port 53` on macOS shows that only one DNS query is sent
	// when testing with `go test -run Timing`, but when simply running `go
	// test`, two queries are sent (and fails the test).
	assert.True(t, lap1.Sub(start) > 10*lap2.Sub(lap1), "Full DNS lookup should be at least 10 times slower than cached lookup")
	fmt.Printf("%v vs %v\n", lap1.Sub(start), lap2.Sub(lap1))
}
