package storage

import(
	dt "GO-DAG/DataTypes"
	// "fmt"
	"GO-DAG/serialize"
	"GO-DAG/Crypto"
	"sync"
)

var OrphanedTransactions = make(map[string]dt.Vertex)
var Mux sync.Mutex 


func AddTransaction(dag *dt.DAG,tx dt.Transaction, signature []byte) bool {

	// change this function for the storage node
	var node dt.Vertex
	var duplicationCheck bool
	duplicationCheck = false
	s := serialize.SerializeData(tx)
	hash := Crypto.Hash(s)
	h := Crypto.EncodeToHex(hash[:])

	dag.Mux.Lock()
	if _,ok := dag.Graph[h] ; !ok{  //Duplication check
		node.Tx = tx
		node.Signature = signature
		left := Crypto.EncodeToHex(tx.LeftTip[:])
		right := Crypto.EncodeToHex(tx.RightTip[:]) 
		l,ok_l := dag.Graph[left]
		r,ok_r := dag.Graph[right]
		if !ok_l || !ok_r {
			if !ok_l {
				OrphanedTransactions[left] = node	
			}
			if !ok_r {
				OrphanedTransactions[right] = node
			}
		} else {
			dag.Graph[h] = node
			if left == right {
				l.Neighbours = append(l.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			} else {
				l.Neighbours = append(l.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
				r.Neighbours = append(r.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])] = r
			}
			duplicationCheck = true
			//fmt.Println("Added Transaction ",h)
		}
	}
	dag.Mux.Unlock()
	if duplicationCheck {
		checkOrphanedTransactions(h,dag)
	}
	return duplicationCheck
}

func checkOrphanedTransactions(h string,dag *dt.DAG) {
	Mux.Lock()
	node,ok := OrphanedTransactions[h]
	Mux.Unlock()
	if ok {
		if AddTransaction(dag,node.Tx,node.Signature) {
			//fmt.Println("resolved Transaction")
		}
	}
	Mux.Lock()
	delete(OrphanedTransactions,h)
	Mux.Unlock()
	return 
}