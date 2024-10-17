package p2p

import "net"

// p2pmessage holds any arbitrary data that is being sent
// over the each transport  between two nodes in the network
type RPC struct {
	from    net.Addr
	Payload []byte
}
