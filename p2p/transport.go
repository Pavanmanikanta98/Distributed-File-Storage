package p2p

import "net"

// Peer is an interface that represents the remote node
// peer embeds net.Conn interface which is basically also implements
// the write and read interfaces
// NOTE: this is peak of golang engineering
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()

	// conn() net.Conn
	// RemoteAddr() net.Addr
	// Close() error
}

// Transport is anything that handles communication btw the nodes in the network
// This can be of the form  ( TCP, UDP, websocket, ...)
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	// ListenAddr() string
}
