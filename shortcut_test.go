package shortcut

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllow(t *testing.T) {
	s := New([]string{"127.0.0.0/24"}, []string{"fe80::1/64", "::/64"})
	assert.True(t, s.Allow("127.0.0.1:8888"))
	assert.True(t, s.Allow("localhost:8888"))
	assert.True(t, s.Allow("localhost"))
	assert.False(t, s.Allow("1.2.4.5:8888"))
	assert.False(t, s.Allow("not-exist.com"))
}
