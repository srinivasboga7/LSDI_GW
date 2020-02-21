package p2p

import (
	dt "GO-DAG/DataTypes"
	"GO-DAG/serialize"
	sh "GO-DAG/sharding"
	"errors"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	haltServer sync.Mutex
)

type handshakeMsg struct {
	ID PeerID
}

func (msg *handshakeMsg) encode() []byte {
	return append(msg.ID.IP, msg.ID.PublicKey...)
}

func validateHandshakeMsg(reply []byte) (PeerID, error) {
	// figure out some validation criteria
	var p PeerID
	p.IP = reply[:4]
	p.PublicKey = reply[4:]
	return p, nil
}

// Server ...
type Server struct {
	peers             []Peer
	maxPeers          uint32
	HostID            PeerID
	ec                chan error
	mux               sync.Mutex
	NewPeer           chan Peer
	BroadcastMsg      chan Msg
	RemovePeer        chan Peer
	ShardingSignal    chan dt.ShardSignal
	ShardTransactions chan dt.ShardTransaction
	// ...
}

// GetRandomPeer ...
func (srv *Server) GetRandomPeer() Peer {
	var p Peer
	for {
		time.Sleep(time.Second)
		srv.mux.Lock()
		if len(srv.peers) > 0 {
			p = srv.peers[0]
			srv.mux.Unlock()
			break
		}
		srv.mux.Unlock()
	}
	return p
}

// setupConn validates a handshake with the other peer
// Adds the new peer to the list of known peers
func (srv *Server) setupConn(conn net.Conn) error {

	msg, err := ReadMsg(conn)
	if err != nil {
		return err
	}

	// validate hanshake message
	pid, err := validateHandshakeMsg(msg.Payload)

	// reply with a proper hanshake
	if msg.ID == 0x00 {
		var hMsg handshakeMsg
		hMsg.ID = srv.HostID
		buf := hMsg.encode()
		var msg Msg
		msg.ID = hsMsg
		msg.LenPayload = uint32(len(buf))
		msg.Payload = buf
		if err := SendMsg(conn, msg); err != nil {
			return err
		}
	} else {
		return errors.New("bad handshake")
	}

	p := newPeer(conn, pid)
	srv.AddPeer(p)
	srv.NewPeer <- *p
	return nil
}

func (srv *Server) performHandshake(c net.Conn, p PeerID) error {
	// define handshake msg

	hmsg := handshakeMsg{srv.HostID}
	buf := hmsg.encode()
	var msg Msg
	msg.ID = hsMsg
	msg.LenPayload = uint32(len(buf))
	msg.Payload = buf

	// sending the handshake msg
	if err := SendMsg(c, msg); err != nil {
		return err
	}
	reply, err := ReadMsg(c)
	if err != nil {
		return err
	}
	// validate the reply figure out
	pid, err := validateHandshakeMsg(reply.Payload)
	if !pid.Equals(p) {
		return errors.New("Invalid Handshake Msg")
	}
	return nil
}

// AddPeer ...
func (srv *Server) AddPeer(p *Peer) {
	srv.mux.Lock()
	srv.peers = append(srv.peers, *p)
	srv.mux.Unlock()
	go p.run()
	return
}

// RemovePeer ...
func (srv *Server) removePeer(peer Peer) {
	// terminate the corresponding go routine and cleanup
	srv.mux.Lock()
	for i, p := range srv.peers {
		if p.ID.Equals(peer.ID) {
			srv.peers[i] = srv.peers[len(srv.peers)-1]
			srv.peers = srv.peers[:len(srv.peers)-1]
			break
		}
	}
	srv.mux.Unlock()
	close(peer.closed)
	return
}

func (srv *Server) listenForConns() {
	listener, _ := net.Listen("tcp", ":8060")
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
	addr := strconv.Itoa(int(b[0])) + "." + strconv.Itoa(int(b[1])) + "."
	addr += strconv.Itoa(int(b[2])) + "." + strconv.Itoa(int(b[3])) + ":8060"
	return addr
}

func (srv *Server) initiateConnection(pID PeerID) (net.Conn, error) {
	conn, err := net.Dial("tcp", parseAddr(pID.IP))
	if err != nil {
		return conn, err
	}

	err = srv.performHandshake(conn, pID)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (srv *Server) discOldPeers() {
	srv.mux.Lock()
	for _, p := range srv.peers {
		if p.ID.ShardID != srv.HostID.ShardID {
			var msg Msg
			msg.ID = 3
			msg.LenPayload = 0
			p.Send(msg)
			close(p.closed)
		}
	}
	srv.mux.Unlock()
}

// Run starts the server
func (srv *Server) Run() {

	srv.ec = make(chan error)
	// start the server
	go srv.listenForConns()
	time.Sleep(time.Second)

	// start the discovery and request peer
	var pIds []PeerID
	pIds = FindPeers(&srv.HostID)
	// iteratively connect with peers
	for _, pID := range pIds {
		// handshake phase
		conn, err := srv.initiateConnection(pID)
		if err != nil {
			log.Println(err)
		} else {
			p := newPeer(conn, pID)
			srv.AddPeer(p)
			srv.NewPeer <- *p
		}
	}

	var tempPeers []PeerID
	go func() {
		for {
			signal := <-srv.ShardingSignal
			haltServer.Lock()
			var sign []byte
			Shardingtx, err := sh.MakeShardingtx(srv.HostID.PublicKey, signal, sign)
			srv.HostID.ShardID = Shardingtx.ShardNo
			if err != nil {
				log.Println(err)
			}
			var msg Msg
			msg.ID = 36
			msg.Payload = serialize.Encode(Shardingtx)
			msg.LenPayload = uint32(len(msg.Payload))
			Send(msg, srv.peers)
			// sign the transaction

			// remove tempPeers with different shardID

			var shPeers []PeerID
			for _, p := range tempPeers {
				if p.ShardID == srv.HostID.ShardID {
					shPeers = append(shPeers, p)
				}
			}

			tempPeers = shPeers

			for {
				// check the number of peers with same shardNo
				if len(tempPeers) > 3 {
					break
				}
				tx := <-srv.ShardTransactions
				if tx.ShardNo == srv.HostID.ShardID {
					var p PeerID
					copy(tx.IP[:], p.IP)
					copy(tx.From[:], p.PublicKey)
					p.ShardID = tx.ShardNo
					tempPeers = append(tempPeers, p)
				}
			}
			// disconect with the other peers
			// connect with these peers with same shardNo
			// check if the routing reaches storage node or some other mechanism

			for _, pID := range tempPeers {
				// check first if it's already present in srv.Peers
				conn, err := srv.initiateConnection(pID)
				if err != nil {
					log.Println(err)
				} else {
					p := newPeer(conn, pID)
					srv.AddPeer(p)
					srv.NewPeer <- *p
				}
			}
			// send the discovery server appropriate information
			updateShardID(srv.HostID)
			srv.discOldPeers()
			haltServer.Unlock()
			var emptySlice []PeerID
			tempPeers = emptySlice
		}
	}()

	for {
		haltServer.Lock()
		select {
		// listen
		case msg := <-srv.BroadcastMsg:
			Send(msg, srv.peers)
		case <-srv.ec:
			log.Fatal("error")
		case p := <-srv.RemovePeer:
			srv.removePeer(p)
		case tx := <-srv.ShardTransactions:
			var p PeerID
			copy(tx.IP[:], p.IP)
			copy(tx.From[:], p.PublicKey)
			p.ShardID = tx.ShardNo
			dup := false
			for _, peer := range tempPeers {
				if peer.Equals(p) {
					dup = true
				}
			}
			if !dup {
				tempPeers = append(tempPeers, p)
			}
		}
		haltServer.Unlock()
	}
}

// Send ...
func Send(msg Msg, peers []Peer) {
	for _, p := range peers {
		var err error
		if !p.ID.Equals(msg.Sender) {
			err = SendMsg(p.rw, msg)
		}
		if err != nil {
			log.Println("problem sending to peer")
		}
	}
}
