package shortcut

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDial(t *testing.T) {
	var directDialer *net.Dialer = &net.Dialer{}

	var dialed string
	proxiedDial := func(ctx context.Context, network, address string) (net.Conn, error) {
		dialed = address
		return nil, errors.New("fail intentionally")
	}

	d := New([]string{"127.0.0.0/24"}, []string{"fe80::1/64", "::/64"}).Dialer(proxiedDial, directDialer.DialContext)
	_, e := d.Dial("tcp", "127.0.0.1:8888")
	assert.Equal(t, "", dialed,
		"should dial directly if address is in the list")
	assert.Error(t, e)
	_, e = d.Dial("tcp", "localhost:8888")
	assert.Equal(t, "", dialed,
		"should dial directly if address is in the list")
	_, e = d.Dial("tcp", "localhost")
	assert.Equal(t, "", dialed,
		"should dial directly if address is in the list")
	_, e = d.Dial("tcp", "1.2.4.5:8888")
	assert.Equal(t, "1.2.4.5:8888", dialed,
		"should dial proxy if address not in the list")
	assert.Error(t, e)
	_, e = d.Dial("tcp", "not-exist.com")
	assert.Equal(t, "not-exist.com", dialed,
		"should dial proxy if address not in the list")
	assert.Error(t, e)
}
