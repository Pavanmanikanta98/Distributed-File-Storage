package p2p

// ----------------------------------------------------------------
// HandShakeFunc ...?
// explaination ...
// HandShakeFunc type is used to represent a function that can be used for a peer to peer handshake process
type HandShakeFunc func(Peer) error

func NoPHandShakeFunc(Peer) error {
	return nil
}
