package storage

import (
	dt "GO-DAG/DataTypes"
	db "GO-DAG/database"
	"GO-DAG/serialize"
	"crypto/sha256"
	"log"
	"sync"
	// "net"
)

var orphanedTransactions = make(map[string][]dt.Vertex)
var mux sync.Mutex

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
	s = append(s, signature...)
	dag.Mux.Lock()
	if _, ok := dag.Graph[h]; !ok { //Duplication check
		node.Tx = tx
		node.Signature = signature
		var tip [32]byte
		if tx.LeftTip == tip && tx.RightTip == tip {
			dag.Genisis = h
			dag.Graph[h] = node
			duplicationCheck = 1
			// db.AddToDb(Txid, s)
		} else {
			left := serialize.EncodeToHex(tx.LeftTip[:])
			right := serialize.EncodeToHex(tx.RightTip[:])
			l, okL := dag.Graph[left]
			r, okR := dag.Graph[right]
			if !okL || !okR {
				log.Println("orphan transaction")
				mux.Lock()
				if !okL {
					orphanedTransactions[left] = append(orphanedTransactions[left], node)
				}
				if !okR {
					orphanedTransactions[right] = append(orphanedTransactions[right], node)
				}
				mux.Unlock()
				duplicationCheck = 2
			} else {
				dag.Graph[h] = node
				if left == right {
					l.Neighbours = append(l.Neighbours, h)
					dag.Graph[serialize.EncodeToHex(tx.LeftTip[:])] = l
				} else {
					l.Neighbours = append(l.Neighbours, h)
					dag.Graph[serialize.EncodeToHex(tx.LeftTip[:])] = l
					r.Neighbours = append(r.Neighbours, h)
					dag.Graph[serialize.EncodeToHex(tx.RightTip[:])] = r
				}
				duplicationCheck = 1
				// db.AddToDb(Txid, s)
			}
		}
	}
	dag.Mux.Unlock()
	if duplicationCheck == 1 {
		checkorphanedTransactions(h, dag, s)
	}
	return duplicationCheck
}

//GetTransactiondb Wrapper function for GetTransaction in db module
func GetTransactiondb(Txid []byte) (dt.Transaction, []byte) {
	stream := db.GetValue(Txid)
	var retval dt.Transaction
	var sig []byte
	retval, sig = serialize.Decode32(stream, uint32(len(stream)))
	return retval, sig
}

//CheckifPresentDb Wrapper function for CheckKey in db module
func CheckifPresentDb(Txid []byte) bool {
	return db.CheckKey(Txid)
}

//GetAllHashes Wrapper function for GetAllKeys in db module
func GetAllHashes() [][]byte {
	return db.GetAllKeys()
}

//checkorphanedTransactions Checks if any other transaction already arrived has any relation with this transaction, Used in the AddTransaction function
func checkorphanedTransactions(h string, dag *dt.DAG, serializedTx []byte) {

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
