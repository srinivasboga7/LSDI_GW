package p2p

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

const (
	hsMsg   = 0x00
	pingMsg = 0x01
	pongMsg = 0x02
	discMsg = 0x03
)

const (
	baseprotoLength = uint32(16)
	pingMsgInterval = 15 * time.Second
)

// PeerID is a structure to identify each unique peer in the P2P network
// using IPv4 Addr and PublicKey
type PeerID struct {
	IP        []byte
	PublicKey []byte
}

// Peer represents a connected remote node
type Peer struct {
	ID     PeerID
	rw     net.Conn
	in     chan Msg // recieves the read msgs
	ec     chan error
	closed chan struct{}
}

func newPeer(c net.Conn, pID PeerID) Peer {
	peer := Peer{
		ID:     pID,
		rw:     c,
		in:     make(chan Msg, 5), // buffered channel to send the msgs
		ec:     make(chan error),
		closed: make(chan struct{}),
	}
	return peer
}

// Send is used for sending the message to the corresponding peer
func (p *Peer) Send(msg Msg) {
	SendMsg(p.rw, msg)
}

func (p *Peer) readLoop(readErr chan error) {
	for {
		msg, err := ReadMsg(p.rw)
		if err != nil {
			select {
			case readErr <- err:
			case <-p.closed:
			}
			return
		}
		go p.handleMsg(msg)
	}
}

// pingLoop is used to check the status of connection
func (p *Peer) pingLoop(e chan error) {
	for {
		time.Sleep(pingMsgInterval)
		var ping Msg
		ping.ID = pingMsg
		if err := SendMsg(p.rw, ping); err != nil {
			select {
			case e <- err:
			case <-p.closed:
			}
			break
		}
	}
}

func (p *Peer) handleMsg(msg Msg) error {
	switch {
	case msg.ID == pingMsg:
		var pong Msg
		pong.ID = pongMsg
		go SendMsg(p.rw, pong)
	case msg.ID == discMsg:
		// close the connection

	case msg.ID < baseprotoLength:
		return errors.New("Invalid msg")
	default:
		select {
		case p.in <- msg:
			// pass on the msg to dag logic
		case <-p.closed:
			return nil
		}
	}
	return nil
}

func (p *Peer) run() {
	// setup the peer and run readLoop and pingLoop
	// wait for errors and terminate the both go routines
	var writeErr chan error
	var readErr chan error
	writeErr = make(chan error)
	readErr = make(chan error)

	var wg sync.WaitGroup
	wg.Add(2)

	go p.pingLoop(writeErr)
	go p.readLoop(readErr)

	select {
	case err := <-readErr:
		log.Println(err)
	case err := <-writeErr:
		log.Println(err)
	case err := <-p.ec:
		log.Println(err)
	}
	close(p.closed)
	wg.Wait()
}
