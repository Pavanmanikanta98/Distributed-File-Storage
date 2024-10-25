package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents the remote node over  a TCP established connection
type TCPPeer struct {
	//The underlying  connection of the peer . which is this case
	// is a TCP connection.
	net.Conn

	//if we dial and  retrive a conn  => outbound == true
	//if we accept and  retrive a conn  => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {

	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
	}

}

func (p *TCPPeer) Send(b []byte) error {

	_, err := p.Conn.Write(b)

	// if err != nil {
	// 	return err
	// }
	// return nil

	return err
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

// Close implements the Transport interface, which will close the   underlying TCP transport connection and disconnect from the server
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transport interface,
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return err
	}
	go t.handleConnection(conn, true)

	return nil

}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	// t.listener = ln
	go t.startAcceptLoop()
	log.Printf("TCP transport listening on %s\t", t.ListenAddr)
	return nil

}

func (t *TCPTransport) startAcceptLoop() {

	for {

		conn, err := t.listener.Accept()

		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Println(" TCP accept error :", err.Error())
			continue
		}

		// fmt.Printf("New Incoming connection %+v\n", conn)

		go t.handleConnection(conn, false)

	}

}

type Temp struct{}

func (t *TCPTransport) handleConnection(conn net.Conn, outbound bool) {

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
