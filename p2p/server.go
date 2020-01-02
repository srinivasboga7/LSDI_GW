package p2p

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"strconv"
	"sync"
)

type handshakeMsg struct {
	ID PeerID
}

func (msg *handshakeMsg) encode() []byte {
	return append(msg.ID.IP, msg.ID.PublicKey...)
}

func validateHandshakeMsg(reply []byte) (*PeerID, error) {
	// figure out some validation criteria
	return nil, nil
}

// Server ...
type Server struct {
	peers    []Peer
	maxPeers uint32
	hostID   PeerID
	ec       chan error
	mux      sync.Mutex
	// ...
}

// setupConn performs a handshake with the other peer
// Adds the new peer to the list of known peers
func (srv *Server) setupConn(conn net.Conn) {
	defer conn.Close()
	pid, err := srv.performHandshake(conn)
	if err != nil {
		// maybe spitout an error or the reason for badhandshake
		log.Println(err, "bad handshake")
		return
	}
	srv.AddPeer(newPeer(conn, *pid))
}

func (srv *Server) performHandshake(c net.Conn) (*PeerID, error) {
	// define handshake msg

	hmsg := handshakeMsg{srv.hostID}
	buf := hmsg.encode()
	var msg Msg
	msg.ID = hsMsg
	msg.LenPayload = uint32(len(buf))
	msg.Payload = buf

	// sending the handshake msg
	if err := SendMsg(c, msg); err != nil {
		return nil, err
	}
	reply, err := ReadMsg(c)
	if err != nil {
		return nil, err
	}
	// validate the reply figure out
	p, err := validateHandshakeMsg(reply.Payload)
	return p, nil
}

// AddPeer ...
func (srv *Server) AddPeer(p Peer) {
	srv.mux.Lock()
	srv.peers = append(srv.peers, p)
	srv.mux.Unlock()
	go p.run()
	return
}

// RemovePeer ...
func (srv *Server) RemovePeer(peer Peer) {
	// terminate the corresponding go routine and cleanup
	srv.mux.Lock()
	for i, p := range srv.peers {
		if bytes.Compare(p.ID.IP, peer.ID.PublicKey) == 0 {
			srv.peers[i] = srv.peers[len(srv.peers)-1]
			srv.peers = srv.peers[:len(srv.peers)-1]
			break
		}
	}
	srv.mux.Unlock()

	return
}

func (srv *Server) listenForConns() {
	listener, _ := net.Listen("tcp", ":8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go srv.setupConn(conn)
	}
}

func parseAddr(b []byte) string {
	var IP struct {
		b1 uint8
		b2 uint8
		b3 uint8
		b4 uint8
	}
	buf := bytes.NewReader(b)
	binary.Read(buf, binary.LittleEndian, &IP)
	addr := strconv.Itoa(int(IP.b1)) + "." + strconv.Itoa(int(IP.b2)) + "."
	addr += strconv.Itoa(int(IP.b3)) + "." + strconv.Itoa(int(IP.b4)) + ":8080"
	return addr
}

func initiateConnection(pID PeerID) (net.Conn, error) {
	conn, err := net.Dial("tcp", parseAddr(pID.IP))
	if err != nil {
		return conn, err
	}

	// wait for the handshake message
	// respond to the handshake message
	msg, err := ReadMsg(conn)
	if err != nil {
		return conn, err
	}

	// validate hanshake message

	// reply with a proper hanshake
	if msg.ID == 0x00 {
		var hMsg handshakeMsg
		hMsg.ID = pID
		if _, err := conn.Write(hMsg.encode()); err != nil {
			return conn, err
		}
	} else {
		return conn, errors.New("bad handshake")
	}

	return conn, nil
}

// New initializes a new node
func New(hID PeerID) *Server {

	// start the server
	var srv Server
	srv.hostID = hID
	srv.ec = make(chan error)
	go srv.listenForConns()

	// start the discovery and request peers
	pIds := FindPeers(hID)

	// iteratively connect with peers
	for _, pID := range pIds {
		// handshake phase
		conn, err := initiateConnection(pID)
		if err != nil {
			conn.Close()
		} else {
			srv.AddPeer(newPeer(conn, pID))
		}
	}

	return &srv
}
