package storage

import(
	dt "GO-DAG/DataTypes"
	"fmt"
	"GO-DAG/serialize"
	"GO-DAG/Crypto"
)

func AddTransaction(dag dt.DAG,tx dt.Transaction) bool {
	
	var node dt.Node
	var duplicationCheck bool
	duplicationCheck = false
	s := serialize.SerializeData(tx)
	hash := Crypto.Hash(s)
	h := Crypto.EncodeToHex(hash[:])
	dag.Mux.Lock()
	if _,ok := dag.Graph[h] ; !ok{  // Duplication check
		node.Tx = tx
		dag.Graph[h] = node  
		l := dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])]
		r := dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])]
		if l.Tx == r.Tx {
			l.Neighbours = append(l.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
		} else {
			l.Neighbours = append(l.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			r.Neighbours = append(r.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])] = r
		}
		//fmt.Println("Succesfully added")
		duplicationCheck = true
		fmt.Println("Added Transaction", h)
	}
	dag.Mux.Unlock()
	return duplicationCheck
}