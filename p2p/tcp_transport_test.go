package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {

	opts := TCPTransportopts{
		ListenAddr:    ":8080",
		HandShakeFunc: NoPHandShakeFunc,
		Decoder:       DefaultDecoder{},
	}
	transport := NewTCPTransport(opts)
	assert.Equal(t, transport.ListenAddr, ":8080")

	// server

	assert.Nil(t, transport.ListenAndAccept())

}
