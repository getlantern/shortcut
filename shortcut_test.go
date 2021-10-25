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
		strings.NewReader("127.0.0.0/24\n8.8.0.0/16\n10.11.0.0/16\n"),
		strings.NewReader("fe80::1/64\n::/64\n2001:4860:4860::8800/120\n"),
		strings.NewReader("10.10.0.0/16\n"),
		strings.NewReader(""),
	)

	allow := func(addr string) Method {
		hit, _ := s.RouteMethod(context.Background(), addr)
		return hit
	}
	assert.Equal(t, allow("127.0.0.1:8888"), Direct)
	assert.Equal(t, allow("localhost:8888"), Direct)
	assert.Equal(t, allow("localhost"), Direct)
	assert.Equal(t, allow("google-public-dns-a.google.com"), Direct)
	assert.Equal(t, allow("google-public-dns-b.google.com"), Direct)
	assert.Equal(t, allow("1.2.4.5:8888"), Unknown)
	assert.Equal(t, allow("1.2.4.5"), Unknown)
	assert.Equal(t, allow("not-exist.com"), Unknown)

	// Test DNS poisoned IP for Iran
	assert.Equal(t, allow("10.10.1.1"), Proxy)

	assert.Equal(t, allow("10.11.1.1"), Direct)
}

func TestContext(t *testing.T) {
	s := NewFromReader(
		strings.NewReader("127.0.0.0/24\n8.8.0.0/16\n"),
		strings.NewReader("fe80::1/64\n::/64\n2001:4860:4860::8800/120\n"),
		strings.NewReader("10.10.0.0/16\n"),
		strings.NewReader(""),
	)
	s.SetResolver(func(ctx context.Context, addr string) (net.IP, error) {
		time.Sleep(100 * time.Millisecond)
		return defaultResolver(ctx, addr)
	})
	ctx := context.Background()
	hit, _ := s.RouteMethod(ctx, "google-public-dns-a.google.com:8888")
	assert.Equal(t, hit, Direct, "host should be allowed when IP is in the list")
	ctx, _ = context.WithTimeout(ctx, 10*time.Millisecond)
	hit, _ = s.RouteMethod(ctx, "google-public-dns-a.google.com:8888")
	assert.Equal(t, hit, Unknown, "host should be disallowed if context exceeded performing DNS lookup")
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
