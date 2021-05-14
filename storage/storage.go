package storage

import (
	dt "LSDI_GW/DataTypes"
	"LSDI_GW/p2p"
	"LSDI_GW/serialize"
	"bytes"
	"crypto/sha256"
	"log"
	"sync"
	"time"
)

var orphanedTransactions = make(map[string][]dt.Vertex)
var mux sync.Mutex

// Server ...
type Server struct {
	ForwardingCh chan p2p.Msg
	ServerCh     chan dt.ForwardTx
	DAG          *dt.DAG
}

//Hash returns the SHA256 hash value
func Hash(b []byte) []byte {
	h := sha256.Sum256(b)
	return []byte(h[:])
}

//AddTransaction checks if transaction if already present in the dag, if not adds to dag and database and returns true else returns false
func AddTransaction(dag *dt.DAG, tx dt.Transaction, signature []byte) int {
	// change this function for the storage node
	var node dt.Vertex
	var duplicationCheck int
	duplicationCheck = 0
	s := serialize.Encode32(tx)
	Txid := Hash(s)
	h := serialize.EncodeToHex(Txid[:])
	dag.Mux.Lock()
	if _, ok := dag.Graph[h]; !ok { //Duplication check
		node.Tx = tx
		node.Signature = signature
		var tip [32]byte
		if tx.LeftTip == tip && tx.RightTip == tip {
			dag.Genisis = h
			dag.Graph[h] = node
			duplicationCheck = 1
			log.Println(tx)
			panic("Found anaother genesis !!!!!!!!")
			// db.AddToDb(Txid, s)
		} else {
			left := serialize.EncodeToHex(tx.LeftTip[:])
			right := serialize.EncodeToHex(tx.RightTip[:])
			l, okL := dag.Graph[left]
			r, okR := dag.Graph[right]
			if !okL || !okR {
				if !okL {
					duplicateOrphanTx := false
					log.Println("left orphan transaction")
					log.Println(left)
					mux.Lock()
					for _, orphantx := range orphanedTransactions[left] {
						if bytes.Compare(orphantx.Signature, node.Signature) == 0 {
							duplicateOrphanTx = true
							break
						}
					}
					if !duplicateOrphanTx {
						orphanedTransactions[left] = append(orphanedTransactions[left], node)
					}
					mux.Unlock()
					duplicationCheck = 2
				}
				if !okR {
					duplicateOrphanTx := false
					log.Println("right orphan transaction")
					log.Println(right)
					mux.Lock()
					for _, orphantx := range orphanedTransactions[right] {
						if bytes.Compare(orphantx.Signature, node.Signature) == 0 {
							duplicateOrphanTx = true
							break
						}
					}
					if !duplicateOrphanTx {
						orphanedTransactions[right] = append(orphanedTransactions[right], node)
					}
					mux.Unlock()
					duplicationCheck = 2
				}
			} else {
				dag.Graph[h] = node
				dag.Length++
				if left == right {
					l.Neighbours = append(l.Neighbours, h)
					dag.Graph[left] = l
				} else {
					l.Neighbours = append(l.Neighbours, h)
					dag.Graph[left] = l
					r.Neighbours = append(r.Neighbours, h)
					dag.Graph[right] = r
				}
				duplicationCheck = 1
				// db.AddToDb(Txid, s)
			}
		}
	}
	dag.Mux.Unlock()
	if duplicationCheck == 1 {
		checkorphanedTransactions(h, dag)
	}
	return duplicationCheck
}

//checkorphanedTransactions Checks if any other transaction already arrived has any relation with this transaction, Used in the AddTransaction function
func checkorphanedTransactions(h string, dag *dt.DAG) {

	mux.Lock()
	list, ok := orphanedTransactions[h]
	mux.Unlock()

	if ok {
		for _, node := range list {
			if AddTransaction(dag, node.Tx, node.Signature) == 1 {
				log.Println("resolved transaction")
			}
		}
	}
	mux.Lock()
	delete(orphanedTransactions, h)
	mux.Unlock()
	return
}

// Run spawns a storage server which spins on a channel (ServerCh) to accept transactions and add them to the DAG.
func (srv *Server) Run() {
	for {
		node := <-srv.ServerCh
		if node.Forward {
			dup := AddTransaction(srv.DAG, node.Tx, node.Signature)
			if dup == 1 {
				var msg p2p.Msg
				msg.ID = 32
				msg.Payload = append(serialize.Encode32(node.Tx), node.Signature...)
				msg.LenPayload = uint32(len(msg.Payload))
				srv.DAG.Mux.Lock()
				if len(srv.DAG.RecentTXs) < 10000 {
					srv.DAG.RecentTXs = append(srv.DAG.RecentTXs, time.Now().UnixNano())
				} else {
					srv.DAG.RecentTXs = srv.DAG.RecentTXs[5000:]
					srv.DAG.RecentTXs = append(srv.DAG.RecentTXs, time.Now().UnixNano())
				}
				srv.DAG.Mux.Unlock()
				srv.ForwardingCh <- msg
			} else if dup == 2 {
				var msg p2p.Msg
				msg.ID = 34
				left := serialize.EncodeToHex(node.Tx.LeftTip[:])
				right := serialize.EncodeToHex(node.Tx.RightTip[:])
				srv.DAG.Mux.Lock()
				if _, t1 := srv.DAG.Graph[left]; !t1 {
					msg.Payload = node.Tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p2p.SendMsg(node.Peer, msg)
				}
				if _, t2 := srv.DAG.Graph[right]; !t2 {
					msg.Payload = node.Tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p2p.SendMsg(node.Peer, msg)
				}
				srv.DAG.Mux.Unlock()
			}
		} else {
			dup := AddTransaction(srv.DAG, node.Tx, node.Signature)
			if dup == 2 {
				var msg p2p.Msg
				msg.ID = 34
				left := serialize.EncodeToHex(node.Tx.LeftTip[:])
				right := serialize.EncodeToHex(node.Tx.RightTip[:])
				srv.DAG.Mux.Lock()
				if _, t1 := srv.DAG.Graph[left]; !t1 {
					msg.Payload = node.Tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p2p.SendMsg(node.Peer, msg)
				}
				if _, t2 := srv.DAG.Graph[right]; !t2 {
					msg.Payload = node.Tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p2p.SendMsg(node.Peer, msg)
				}
				srv.DAG.Mux.Unlock()
			}
		}
	}
}
