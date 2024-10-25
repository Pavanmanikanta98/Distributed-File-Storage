package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/pavanmanikanta98/dfs-with-go/p2p"
)

type FileServerOpts struct {
	// ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
	// TCPTransportOpts  p2p.TCPTransportopts
}

type FileServer struct {
	FileServerOpts
	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	store    *store
	quitch   chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts, // Ensure that opts are stored
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Payload struct {
	key  string
	data []byte
}

// BIG thing to understand here
func (s *FileServer) broadcast(p Payload) error {

	peers := []io.Writer{}

	for _, peer := range s.peers {

		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)

	// return nil
}

// 1.store the file to disk
// 2. broadcast this file to all the known peers in the network
func (s *FileServer) StoreData(key string, r io.Reader) error {

	return nil
}

func (s *FileServer) Stop() {

	close(s.quitch)
}

func (s *FileServer) OnPeer(peer p2p.Peer) error {
	s.peerLock.Lock()

	defer s.peerLock.Unlock()

	s.peers[peer.RemoteAddr().String()] = peer

	log.Printf("connected with (remote) peer %s", peer.RemoteAddr())

	return nil

}

func (s *FileServer) loop() {

	defer func() {
		log.Println("file server stopped due to user quit action")
		s.Transport.Close()

	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)

		case <-s.quitch:
			return

		}
	}

}

func (s *FileServer) BootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {

		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Println("attempting to connect with remote : ", addr)

			if err := s.Transport.Dial(addr); err != nil {

				log.Printf("Failed to dial:  %s: %v\n", addr, err)

			}
		}(addr)

	}

	return nil
}

func (s *FileServer) Start() error {

	if err := s.Transport.ListenAndAccept(); err != nil {

		return err

	}

	s.BootstrapNetwork()

	s.loop()
	fmt.Println("File server died")
	return nil

}

// start with get the mesage and store the message
