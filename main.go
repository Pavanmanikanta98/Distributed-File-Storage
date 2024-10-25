package main

import (
	"bytes"
	"log"

	"github.com/pavanmanikanta98/dfs-with-go/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcptransportOpts := p2p.TCPTransportopts{
		ListenAddr:    listenAddr,
		HandShakeFunc: p2p.NoPHandShakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO : onPeer func
	}

	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)
	if tcpTransport == nil {
		log.Fatal("Failed to initialize TCPTransport")
	}

	// Set up file server options
	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	// Create the file server
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
func main() {

	s1 := makeServer(":4000", "")
	s2 := makeServer(":3000", ":4000")

	go func() {
		log.Fatal(s1.Start())
	}()

	s2.Start()

	data := bytes.NewReader([]byte("My big data file "))

	s2.StoreData("myprivatedata", data)

	// tcptransportOpts := p2p.TCPTransportopts{
	// 	ListenAddr:    ":5000",
	// 	HandShakeFunc: p2p.NoPHandShakeFunc,
	// 	Decoder:       p2p.DefaultDecoder{},
	// 	// TODO : onPeer func
	// }

	// tcpTransport := p2p.NewTCPTransport(tcptransportOpts)
	// if tcpTransport == nil {
	// 	log.Fatal("Failed to initialize TCPTransport")
	// }

	// // Set up file server options
	// fileServerOpts := FileServerOpts{
	// 	StorageRoot:       "5000_network",
	// 	PathTransformFunc: CASPathTransformFunc,
	// 	Transport:         tcpTransport,
	// 	BootstrapNodes:    []string{":4000"},
	// }

	// Create the file server
	// s := NewFileServer(fileServerOpts)

	// go func() {
	// 	time.Sleep(time.Second * 3)
	// 	s.Stop()
	// }()

	// Start the file server
	// if err := s.Start(); err != nil {
	// 	log.Fatalf("Failed to start file server: %v", err)
	// }

	// Keep the main function running
	// select {}

}

// NO longer needed..

// func OnPeer(peer p2p.Peer) error {

// 	fmt.Println("DoIng some Logic with the peer  ouside TCPTransport")
// 	peer.Close()
// 	return nil
// }
// func main() {

// 	tcpOpts := p2p.TCPTransportopts{
// 		ListenAddr:    ":8081",
// 		HandShakeFunc: p2p.NoPHandShakeFunc,
// 		Decoder:       p2p.DefaultDecoder{},
// 		OnPeer:        OnPeer,
// 	}

// 	tr := p2p.NewTCPTransport(tcpOpts)

// 	go func() {
// 		for {
// 			msg := <-tr.Consume()
// 			log.Printf("Received %+v \n", msg)
// 		}
// 	}()
// 	// log.Fatal(tr.ListenAndAccept())
// 	if err := tr.ListenAndAccept(); err != nil {
// 		log.Fatal(err)
// 	}

// 	select {}

// 	// fmt.Println("welcome to distributed file storage")

// }
