package node

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"log"
)

// New ...
func New(hostID p2p.PeerID, dag *dt.DAG) chan p2p.Msg {

	var srv p2p.Server
	srv.HostID = hostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	srv.NewPeer = make(chan p2p.Peer)
	srv.RemovePeer = make(chan p2p.Peer)
	go srv.Run()

	go func() {
		for {
			p := <-srv.NewPeer
			go handle(&p, srv.BroadcastMsg, dag, srv.RemovePeer)
		}
	}()
	return srv.BroadcastMsg
}

// NewBootstrap ...
func NewBootstrap(hostID p2p.PeerID, dag *dt.DAG) chan p2p.Msg {
	var srv p2p.Server
	srv.HostID = hostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	srv.NewPeer = make(chan p2p.Peer)
	srv.RemovePeer = make(chan p2p.Peer)
	go srv.Run()

	go func() {
		for {
			p := <-srv.NewPeer
			go handle(&p, srv.BroadcastMsg, dag, srv.RemovePeer)
		}
	}()
	return srv.BroadcastMsg
}

func handleMsg(msg p2p.Msg, send chan p2p.Msg, dag *dt.DAG, p *p2p.Peer) {
	// check for transactions or request for transactions
	if msg.ID == 32 {
		// transaction
		tx, sign := serialize.DeserializeTransaction(msg.Payload, msg.LenPayload)
		log.Println(Crypto.EncodeToHex(tx.LeftTip[:]))
		if validTransaction(tx, sign) {
			tr := storage.AddTransaction(dag, tx, sign)
			if tr == 1 {
				send <- msg
			} else if tr == 2 {
				var msg p2p.Msg
				msg.ID = 34
				if !storage.CheckifPresentDb(tx.LeftTip[:]) {
					msg.Payload = tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
				if !storage.CheckifPresentDb(tx.RightTip[:]) {
					msg.Payload = tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
			}
		}
	} else if msg.ID == 34 {
		// request for transaction
		hash := msg.Payload
		v := dag.Graph[Crypto.EncodeToHex(hash)]
		tx := v.Tx
		sign := v.Signature
		var respMsg p2p.Msg
		respMsg.ID = 33
		respMsg.Payload = append(serialize.Encode(tx), sign...)
		respMsg.LenPayload = uint32(len(respMsg.Payload))
		p.Send(respMsg)
	} else if msg.ID == 33 {
		// transaction recieved after request
		tx, sign := serialize.DeserializeTransaction(msg.Payload, msg.LenPayload)
		log.Println(Crypto.EncodeToHex(tx.LeftTip[:]))
		if validTransaction(tx, sign) {
			tr := storage.AddTransaction(dag, tx, sign)
			if tr == 2 {
				var msg p2p.Msg
				msg.ID = 34
				if !storage.CheckifPresentDb(tx.LeftTip[:]) {
					msg.Payload = tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
				if !storage.CheckifPresentDb(tx.RightTip[:]) {
					msg.Payload = tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
			}
		}
	} else if msg.ID == 35 {
		// sharding signal

	} else if msg.ID == 36 {
		// sharding transaction
	}
}

// read the messages and handle
func handle(p *p2p.Peer, send chan p2p.Msg, dag *dt.DAG, errChan chan p2p.Peer) {
	for {
		msg, err := p.GetMsg()
		if err != nil {
			errChan <- *p
			log.Println(err)
			break
		}
		handleMsg(msg, send, dag, p)
	}
}

func validTransaction(tx dt.Transaction, sign []byte) bool {
	return true
}

func getAllKeys(dag *dt.DAG) []string {

	var hashes []string
	dag.Mux.Lock()

	for k := range dag.Graph {
		hashes = append(hashes, k)
	}

	dag.Mux.Unlock()
	return hashes
}
