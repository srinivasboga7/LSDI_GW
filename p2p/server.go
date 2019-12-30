package p2p

import (
	"log"
	"net"
)

// Server ...
type Server struct {
	peers []Peer
}

// setupConn performs a handshake with the other peer
// Adds the new peer to the list of known peers
func setupConn(conn net.Conn) {
	defer conn.Close()

}

func performHandshake(net.Conn) {
	// define handshake msg and proceed with this function
}

// AddPeer ...
func (srv *Server) AddPeer(p Peer) {

}

// RemovePeer ...
func (srv *Server) RemovePeer(p Peer) {
	// terminate the go routine and cleanup the list
}

func listenForConns() {
	listener, _ := net.Listen("tcp", ":8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go setupConn(conn)
	}
}
