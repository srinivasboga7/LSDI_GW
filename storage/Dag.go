package storage

import(
	dt "GO-DAG/DataTypes"
	"fmt"
	"GO-DAG/serialize"
	"GO-DAG/Crypto"
	//"GO-DAG/consensus"
)

func AddTransaction(dag *dt.DAG,tx dt.Transaction, signature []byte) bool {

	var node dt.Node
	var duplicationCheck bool
	duplicationCheck = false
	s := serialize.SerializeData(tx)
	hash := Crypto.Hash(s)
	h := Crypto.EncodeToHex(hash[:])
	dag.Mux.Lock()
	if _,ok := dag.Graph[h] ; !ok{  // Duplication check
		node.Tx = tx
		node.Signature = signature
		dag.Graph[h] = node  
		l,ok := dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])]
		if !ok {
			fmt.Println("something wrong")
		}
		r,ok := dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])]
		if !ok {
			fmt.Println("something wrong")
		}
		if l.Tx == r.Tx {
			l.Neighbours = append(l.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
		} else {
			l.Neighbours = append(l.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			r.Neighbours = append(r.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])] = r
		}
		duplicationCheck = true
		//fmt.Println("Added Transaction", h)
	}
	dag.Mux.Unlock()
	return duplicationCheck
}