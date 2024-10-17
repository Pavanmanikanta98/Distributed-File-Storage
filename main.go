package main

import (
	"fmt"
	"log"

	"github.com/pavanmanikanta98/dfs-with-go/p2p"
)

func OnPeer(peer p2p.Peer) error {

	fmt.Println("DoIng some Logic with the peer  ouside TCPTransport")
	peer.Close()
	return nil
}
func main() {

	tcpOpts := p2p.TCPTransportopts{
		ListenAddr:    ":8081",
		HandShakeFunc: p2p.NoPHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			log.Printf("Received %+v \n", msg)
		}
	}()
	// log.Fatal(tr.ListenAndAccept())
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

	// fmt.Println("welcome to distributed file storage")

}
