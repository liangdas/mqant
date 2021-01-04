package iptool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsInnerIp(t *testing.T) {
	assert.Equal(t, IsInnerIp("192.168.0.2"), true)
	assert.Equal(t, IsInnerIp("172.17.0.1"), true)
	assert.Equal(t, IsInnerIp("220.249.15.134"), false)
	assert.Equal(t, IsInnerIp("100.122.245.119"), true)

	gip := GetGlobalIPFromXforwardedFor("192.168.0.2, 100.64.235.3, 202.106.9.134, 101.32.47.171, 34.102.170.30, 127.0.0.1")
	assert.Equal(t, gip, "202.106.9.134")
}
