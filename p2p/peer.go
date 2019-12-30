package p2p

import (
	"errors"
	"net"
)

const (
	pingMsg = 0x01
	pongMsg = 0x02
	discMsg = 0x03
)

const (
	baseprotoLength = uint32(16)
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
	e      chan error
	closed chan struct{}
}

func newPeer(c net.Conn, pID PeerID) Peer {
	peer := Peer{
		ID:     pID,
		rw:     c,
		in:     make(chan Msg, 5), // buffered channel to send the msgs
		e:      make(chan error),
		closed: make(chan struct{}),
	}
	return peer
}

func (p *Peer) readLoop() {
	for {
		msg, err := ReadMsg(p.rw)
		if err != nil {
			p.e <- err
			return
		}
		p.handleMsg(msg)
	}
}

func (p *Peer) handleMsg(msg Msg) error {
	switch {
	case msg.ID == pingMsg:
		var pong Msg
		pong.ID = pongMsg
		go SendMsg(p.rw, pong, p.e)
	case msg.ID == discMsg:
		// close the connection

	case msg.ID < baseprotoLength:
		return errors.New("Invalid msg")
	default:
		select {
		case p.in <- msg:

		}
	}
	return nil
}

func (p *Peer) run() error {

	return nil
}
