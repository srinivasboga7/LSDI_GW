package storage

import(
	dt "GO-DAG/DataTypes"
	"fmt"
	"GO-DAG/serialize"
	"GO-DAG/Crypto"
)

func AddTransaction(dag dt.DAG,c chan dt.Transaction) {
	for {
		var node dt.Node
		node.Tx = <-c
		s := serialize.SerializeData(node.Tx)
		hash := Crypto.Hash(s)
		h := Crypto.EncodeToHex(hash[:])
		dag.Graph[h] = node
		fmt.Println("Succesfully added")
	}
}