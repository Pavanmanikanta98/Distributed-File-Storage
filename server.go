package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/pavanmanikanta98/dfs-with-go/p2p"
)

type FileServerOpts struct {
	EncKey            []byte
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
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

// BIG thing to understand here
func (s *FileServer) stream(msg *Message) error {

	//Uses io.MultiWriter to broadcast to all peers at once.
	//If any single peer connection fails,io.MultiWriter stops,
	//causing the entire broadcast operation to fail for all peers.

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
		return fmt.Errorf("failed to encode payload: %v", err)
	}

	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	for addr, peer := range s.peers {
		if _, err := io.Copy(peer, bytes.NewReader(buf.Bytes())); err != nil {
			log.Printf("Error broadcasting to peer %s: %v", addr, err)
			delete(s.peers, addr)
		}
	}

	return nil
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		log.Printf("Error encoding message: %v", err) // Log the error
		return fmt.Errorf("failed to encode payload: %v", err)
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err

		}

	}

	return nil

}

type Message struct {
	// From    string
	Payload any
}

type MessageStorageFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) GET(key string) (io.Reader, error) {

	if s.store.Has(key) {

		fmt.Printf("[%s] Serving File (%s) found locally. Reading from disk...\n", s.Transport.Addr(), key)
		_, file, err := s.store.Read(key)
		return file, err
	}

	fmt.Printf("[%s] Don't have file (%s )locally, fetching from network... \n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			Key: hashKey(key),
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		// First read the file size so we can limit the amount of bytes
		//that we read from the connection, so it will  not hanging.
		var fileSize int64
		err := binary.Read(peer, binary.LittleEndian, &fileSize)

		if err != nil {
			continue
		}

		n, err := s.store.WriteDecrypt(s.EncKey, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}

		// n, err := s.store.Write(key, io.LimitReader(peer, fileSize))
		// if err != nil {
		// 	//continue
		// 	return nil, err
		// }

		fmt.Printf("[%s] received (%d) bytes  over the network from (%s)\n", s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	_, file, err := s.store.Read(key)

	return file, err
}

func (s *FileServer) Store(key string, r io.Reader) error {
	// 1.store the file to disk
	// 2. broadcast this file to all the known peers in the network

	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStorageFile{
			Key:  hashKey(key),
			Size: size + 16,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 5)

	// MultiWriter staff

	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	n, err := copyEncrypt(s.EncKey, fileBuffer, mw)

	if err != nil {
		return err
	}

	// for _, peer := range s.peers {

	// 	peer.Send([]byte{p2p.IncomingStream})
	// 	n, err := copyEncrypt(s.EncKey, fileBuffer, peer)

	// 	if err != nil {
	// 		return err
	// 	}

	// 	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.Addr(), n)

	// }

	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.Addr(), n)

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
		log.Println("File server stopped due to Error or user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				if err == io.EOF {
					log.Println("Received EOF, continuing to next message.")
					continue
				}
				log.Printf("Error decoding Message: %v", err)
				continue
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Printf("Error handling message: %v", err)
				continue
			}

			// fmt.Printf("%+v\n", msg.Payload)

		case <-s.quitch:
			return
		}
	}
}

// handle the payload or Message
func (s *FileServer) handleMessage(from string, msg *Message) error {

	switch v := msg.Payload.(type) {
	case MessageStorageFile:
		return s.handleMessageStoreFile(from, v)
		// fmt.Printf("Received Message : %+v\n", v)

	case MessageGetFile:
		return s.handleMessageFile(from, v)
	default:
		log.Printf("Unhandled payload type: %T", v)
		return nil
	}

	// return nil
}

func (s *FileServer) handleMessageFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it does not found in the disk", s.Transport.Addr(), msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network \n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	// NEW LEARNING :  In GO is basically you could assert if
	//certain implementations are true like you would say that
	//ReadCloser , a boolean equals the reader is that a readCloser
	//(if it's a read closer) then  ok will be true next follows the next
	if rc, ok := r.(io.ReadCloser); ok {

		defer rc.Close()
	}

	peer, ok := s.peers[from]

	if !ok {
		return fmt.Errorf("peer %s not found in the peer list", from)
	}

	// First send the "incomingStream" byte to the peer and then we can
	//send the file size as an int64.
	peer.Send([]byte{p2p.IncomingStream})
	// var fileSize int64 = 17
	binary.Write(peer, binary.LittleEndian, fileSize)

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStorageFile) error {

	peer, ok := s.peers[from]

	if !ok {
		return fmt.Errorf("peer %s not found in the peer list", from)
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Printf("[%s]  Writtten %d byte to disk.\n", s.Transport.Addr(), n)
	// peer.(*p2p.TCPPeer).Wg.Done()
	peer.CloseStream()

	return nil
}

func (s *FileServer) BootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {

		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Printf("[%s] attempting to connect with remote %s\n ", s.Transport.Addr(), addr)

			if err := s.Transport.Dial(addr); err != nil {

				log.Printf("[%s] Failed to dial:  %s: %v\n", s.Transport.Addr(), addr, err)

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

func init() {

	// fmt.Println("rtc")
	gob.Register(MessageGetFile{})
	gob.Register(MessageStorageFile{})

}
