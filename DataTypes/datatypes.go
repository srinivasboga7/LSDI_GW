package DataTypes


import(
	"sync"
	"net"
)


type Transaction struct {
	// Definition of data
	Timestamp int64
	Value float64 //could be a string but have to figure out serialization
	From [65]byte //length of public key 33(compressed) or 65(uncompressed)
	Txid [16]byte
	LeftTip [32]byte
	RightTip [32]byte
	Nonce uint32 //temporary based on type of PoW
}

type Peers struct {
	Mux sync.Mutex
	Fds map[string] net.Conn
}

type Vertex struct {
	Tx Transaction
	Signature []byte
	Neighbours [] string // pointers to the neighbours which gets updated when this node is chosen as tip

}

type DAG struct {
	Mux sync.Mutex
	/*
	ChTx chan dt.Node
	ChDuplicateCheck chan string
	ChTipSelection chan string
	*/
	Genisis string
	Graph map[string] Vertex // string is the hash of the transaction(Node.Tx)
}

type Request struct {
	Data []byte
	IP string
	Conn net.Conn
}