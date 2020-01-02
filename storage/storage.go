package storage

import(
	dt "GO-DAG/DataTypes"
	"GO-DAG/serialize"
	db "GO-DAG/database"
	"crypto/sha256"
	"sync"
	// "net"
)

var OrphanedTransactions = make(map[string] []dt.Vertex)
var Mux sync.Mutex 

//Hash returns the SHA256 hash value
func Hash(b []byte) []byte {
	h := sha256.Sum256(b)
	return []byte(h[:])
}

//AddTransaction checks if transaction if already present in the dag, if not adds to dag and database and returns true else returns false
func AddTransaction(dag *dt.DAG,tx dt.Transaction, signature []byte, serializedTx []byte, ) int {

	// change this function for the storage node
	var node dt.Vertex
	var duplicationCheck int
	duplicationCheck = 0
	s := serialize.SerializeData(tx)
	Txid := Hash(s)
	h := serialize.EncodeToHex(Txid[:])
	dag.Mux.Lock()
	if _,ok := dag.Graph[h] ; !ok{  //Duplication check
		node.Tx = tx
		node.Signature = signature
		var tip [32]byte
		if tx.LeftTip == tip && tx.RightTip == tip {
			dag.Genisis = h
			dag.Graph[h] = node
			duplicationCheck = 1
			db.AddToDb(Txid,serializedTx)
		} else {
			left := serialize.EncodeToHex(tx.LeftTip[:])
			right := serialize.EncodeToHex(tx.RightTip[:]) 
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
					dag.Graph[serialize.EncodeToHex(tx.LeftTip[:])] = l
				} else {
					l.Neighbours = append(l.Neighbours,h)
					dag.Graph[serialize.EncodeToHex(tx.LeftTip[:])] = l
					r.Neighbours = append(r.Neighbours,h)
					dag.Graph[serialize.EncodeToHex(tx.RightTip[:])] = r
				}
				duplicationCheck = 1
				db.AddToDb(Txid,serializedTx)
			}
		}
	}
	dag.Mux.Unlock()
	if duplicationCheck == 1{
		checkOrphanedTransactions(h,dag,serializedTx)
	}
	return duplicationCheck
}

//GetTransactiondb Wrapper function for GetTransaction in db module
func GetTransactiondb(Txid []byte) (dt.Transaction,[]byte) {
	stream := db.GetValue(Txid)
	return serialize.DeserializeTransaction(stream)
}

//checkifPresentDb Wrapper function for CheckKey in db module
func checkifPresentDb(Txid []byte) bool {
	return db.CheckKey(Txid)
}

//GetAllHashes Wrapper function for GetAllKeys in db module
func GetAllHashes() [][]byte{
	return db.GetAllKeys()
}

//checkOrphanedTransactions Checks if any other transaction already arrived has any relation with this transaction, Used in the AddTransaction function
func checkOrphanedTransactions(h string,dag *dt.DAG, serializedTx []byte) {
	Mux.Lock()
	list,ok := OrphanedTransactions[h]
	Mux.Unlock()
	if ok {
		for _,node := range list {
			if AddTransaction(dag,node.Tx,node.Signature,serializedTx) == 1{
				// fmt.Println("resolved Transaction")
			}
		}
	}
	Mux.Lock()
	delete(OrphanedTransactions,h)
	Mux.Unlock()
	return 
}