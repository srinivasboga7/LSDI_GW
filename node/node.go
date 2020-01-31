package node

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"GO-DAG/sync"
	"encoding/json"
	"log"
)

// New ...
func New(hostID p2p.PeerID, dag *dt.DAG) chan p2p.Msg {

	var srv p2p.Server
	srv.HostID = hostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	go srv.Run()

	go func() {
		for {
			p := <-srv.NewPeer
			go handle(p, srv.BroadcastMsg, dag, srv.RemovePeer)
		}
	}()
	sync.Sync(dag, srv.GetRandomPeer())
	return srv.BroadcastMsg
}

// NewBootstrap ...
func NewBootstrap(hostID p2p.PeerID, dag *dt.DAG) chan p2p.Msg {
	var srv p2p.Server
	srv.HostID = hostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	go srv.Run()

	go func() {
		for {
			p := <-srv.NewPeer
			go handle(p, srv.BroadcastMsg, dag, srv.RemovePeer)
		}
	}()
	return srv.BroadcastMsg
}

func handleMsg(msg p2p.Msg, send chan p2p.Msg, dag *dt.DAG, p p2p.Peer) {
	// check for transactions or request for transactions
	if msg.ID == 32 {
		// transaction
		tx, sign := serialize.Decode(msg.Payload, msg.LenPayload)
		if validTransaction(tx, sign) {
			if storage.AddTransaction(dag, tx, sign) != 0 {
				send <- msg
			}
		}
	} else if msg.ID == 34 {
		hash := msg.Payload
		v := dag.Graph[Crypto.EncodeToHex(hash)]
		tx := v.Tx
		sign := v.Signature
		var msg p2p.Msg
		msg.ID = 32
		msg.Payload = append(serialize.Encode(tx), sign...)
		msg.LenPayload = uint32(len(msg.Payload))
		p.Send(msg)
	} else if msg.ID == 33 {
		txHashes := getAllKeys(dag)
		msg.ID = 33
		msg.Payload, _ = json.Marshal(txHashes)
		msg.LenPayload = uint32(len(msg.Payload))
		p.Send(msg)
	}
}

// read the messages and handle
func handle(p p2p.Peer, send chan p2p.Msg, dag *dt.DAG, errChan chan p2p.Peer) {
	for {
		msg, err := p.GetMsg()
		if err != nil {
			errChan <- p
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
