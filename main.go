package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

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

	// go func() {
	// 	for {
	// 		msg := <-tcpTransport.Consume()
	// 		log.Printf("Received %+v \n", msg.Payload)
	// 	}
	// }()

	// Set up file server options
	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionkey(),
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
	s3 := makeServer(":5000", ":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(500 * time.Millisecond)

	go func() {
		log.Fatal(s2.Start())

	}()

	time.Sleep(1 * time.Second)

	go s3.Start()
	time.Sleep(1 * time.Second)

	for i := 0; i < 17; i++ {

		// key := "golang.jpg"
		key := fmt.Sprintf("picture_%d.png", i)

		data := bytes.NewReader([]byte("My big data file... "))

		s3.Store(key, data)

		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.GET(key)
		if err != nil {
			log.Fatal(err)
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			log.Fatal(err)
		}

		fmt.Println(buf.String())

	}

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
