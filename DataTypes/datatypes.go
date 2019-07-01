package DataTypes


import(
	"sync"
	"net"
)


type Transaction struct {
	// Definition of data
	Timestamp int64
	Value float64 //could be a string but have to figure out serialization
	From [N]byte // N - length of public key 33(compressed) or 65(uncompressed)
	LeftTip [32]byte
	RightTip [32]byte
	Nonce uint32 //temporary based on type of PoW
}

type Peers struct {
	Mux sync.Mutex
	Fds map[string] net.Conn
}

type Node struct {
	Tx Transaction
	Weight uint32
	Neighbours [] *Node //pointers to the neighbours
}

type DAG struct {
	Mux sync.Mutex
	Graph map[string] Node // string is the hash of the transaction
}