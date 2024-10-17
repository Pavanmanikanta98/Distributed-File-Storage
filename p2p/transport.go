package p2p

// Peer is an interface that represents the remote node
type Peer interface {
	Close() error
}

// Transport is anything that handles communication btw the nodes in the network
// This can be of the form  ( TCP, UDP, websocket, ...)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}
