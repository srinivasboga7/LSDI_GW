package storage

import(
	dt "GO-DAG/DataTypes"
	// "fmt"
	"GO-DAG/serialize"
	"GO-DAG/Crypto"
	"sync"
)

var OrphanedTransactions = make(map[string] []dt.Vertex)
var Mux sync.Mutex 


func AddTransaction(dag *dt.DAG,tx dt.Transaction, signature []byte) int {

	// change this function for the storage node
	var node dt.Vertex
	var duplicationCheck int
	duplicationCheck = 0
	s := serialize.SerializeData(tx)
	hash := Crypto.Hash(s)
	h := Crypto.EncodeToHex(hash[:])

	dag.Mux.Lock()
	if _,ok := dag.Graph[h] ; !ok{  //Duplication check
		node.Tx = tx
		node.Signature = signature
		var tip [32]byte
		if tx.LeftTip == tip && tx.RightTip == tip {
			dag.Genisis = h
			dag.Graph[h] = node
			duplicationCheck = 1
		} else {
			left := Crypto.EncodeToHex(tx.LeftTip[:])
			right := Crypto.EncodeToHex(tx.RightTip[:]) 
			l,okL := dag.Graph[left]
			r,okR := dag.Graph[right]
			if !okL || !okR {
				if !okL {
					OrphanedTransactions[left] = append(OrphanedTransactions[left],node)
					// fmt.Println("Orphan Transaction")	
				}
				if !okR {
					OrphanedTransactions[right] = append(OrphanedTransactions[right],node)
					// fmt.Println("Orphan Transaction")
				}
				duplicationCheck = 2
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
				duplicationCheck = 1
			}
		}
	}
	dag.Mux.Unlock()
	if duplicationCheck == 1{
		checkOrphanedTransactions(h,dag)
	}
	return duplicationCheck
}

func checkOrphanedTransactions(h string,dag *dt.DAG) {
	Mux.Lock()
	list,ok := OrphanedTransactions[h]
	Mux.Unlock()
	if ok {
		for _,node := range list {
			if AddTransaction(dag,node.Tx,node.Signature) == 1{
				// fmt.Println("resolved Transaction")
			}
		}
	}
	Mux.Lock()
	delete(OrphanedTransactions,h)
	Mux.Unlock()
	return 
}