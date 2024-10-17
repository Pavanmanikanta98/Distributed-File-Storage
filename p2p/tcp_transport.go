package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents the remote node over  a TCP established connection
type TCPPeer struct {
	//conn is the underlying TCP connection of the peer
	conn net.Conn

	//if we dial and  retrive a conn  => outbound == true
	//if we accept and  retrive a conn  => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {

	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}

}

// close implements the peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportopts struct {
	ListenAddr    string
	HandShakeFunc HandShakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportopts
	listener net.Listener
	rpcch    chan RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportopts) *TCPTransport {

	return &TCPTransport{
		TCPTransportopts: opts,
		rpcch:            make(chan RPC),
	}

}

// Consume implements the Transport interface, which wil return readOnly
// channel for Reading the incoming messages received from Another peer in the network .
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	// t.listener = ln
	go t.startAcceptLoop()
	return nil

}

func (t *TCPTransport) startAcceptLoop() {

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		fmt.Printf("New Incoming connection %+v\n", conn)

		go t.handleConnection(conn)

	}

}

type Temp struct{}

func (t *TCPTransport) handleConnection(conn net.Conn) {

	var err error

	defer func() {
		fmt.Printf("Dropping Peer connection %s \n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err = t.HandShakeFunc(peer); err != nil {

		return

	}

	if t.OnPeer != nil {

		if err = t.OnPeer(peer); err != nil {

			return

		}

	}

	// lenDecodeError := 0
	//Read loop

	// for {
	// 	if err := t.Decoder.Decode(conn, msg); err != nil {
	// 		lenDecodeError++
	// 		if lenDecodeError >= 1 {
	// 			fmt.Printf("Too many decode errors, closing connection: %+v\n", peer)
	// 			fmt.Printf("TCP error : %s\n", err)
	// 			conn.Close()
	// 			return
	// 		}
	// 		continue
	// 	}
	rpc := RPC{}
	for {
		if err = t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("TCP error : %s\n", err)
			return

		}

		rpc.from = conn.RemoteAddr()

		t.rpcch <- rpc

		// fmt.Printf("message %+v from %+v\n", string(rpc.Payload), rpc.from.String())
	}

	// fmt.Printf("New Incoming connection %+v\n", peer)

}
