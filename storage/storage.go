package storage

import(
	dt "GO-DAG/DataTypes"
	// "fmt"
	"GO-DAG/serialize"
	"GO-DAG/Crypto"
	db "GO-DAG/database"
	"sync"
	"net"
)

// Transaction DS Transaction structure
type Transaction struct {
	Timestamp int64 // 8 bytes
	Hash [32]byte //could be a string but have to figure out serialization
	From [65]byte //length of public key 33(compressed) or 65(uncompressed)
	LeftTip [32]byte
	RightTip [32]byte
	Nonce uint32 // 4 bytes
}

// Peers maintains the list of all peers connected to the node
type Peers struct {
	Mux sync.Mutex
	Fds map[string] net.Conn
}

// Vertex is a wrapper struct of Transaction 
type Vertex struct {
	Tx Transaction
	Signature []byte
	Neighbours [] string 
}

// DAG defines the data structure to store the blockchain
type DAG struct {
	Mux sync.Mutex
	Genisis string
	Graph map[string] Vertex
}

var OrphanedTransactions = make(map[string] []Vertex)
var Mux sync.Mutex 

//AddTransaction checks if transaction if already present in the dag, if not adds to dag and database and returns true else returns false
func AddTransaction(dag *DAG,tx Transaction, signature []byte, serializedTx []byte) int {

	// change this function for the storage node
	var node Vertex
	var duplicationCheck int
	duplicationCheck = 0
	s := serialize.SerializeData(tx)
	Txid := Crypto.Hash(s)
	h := Crypto.EncodeToHex(Txid[:])
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
				db.AddToDb(h,serializedTx)
			}
		}
	}
	dag.Mux.Unlock()
	if duplicationCheck == 1{
		checkOrphanedTransactions(h,dag)
	}
	return duplicationCheck
}

//checkOrphanedTransactions Checks if any other transaction already arrived has any relation with this transaction, Used in the AddTransaction function
func checkOrphanedTransactions(h string,dag *DAG) {
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