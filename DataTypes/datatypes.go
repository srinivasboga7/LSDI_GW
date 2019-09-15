package DataTypes


import(
	"sync"
	"net"
)

// Transaction defines the structure of the transaction in the blockchain
type Transaction struct {
	Timestamp int64 // 8 bytes
	Hash [32]byte //could be a string but have to figure out serialization
	From [65]byte //length of public key 33(compressed) or 65(uncompressed)
	TxID [16]byte
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