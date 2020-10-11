package DataTypes

import (
	"net"
	"sync"
)

// Transaction defines the structure of the transaction in the blockchain
type Transaction struct {
	Timestamp int64         //8 bytes
	TxType    int32         //Allocate, Revoke, Update, Path announce, Path withdraw
	Msg       BGPMssgHolder //could be a string but have to figure out serialization
	From      [65]byte      //length of public key 33(compressed) or 65(uncompressed)
	LeftTip   [32]byte
	RightTip  [32]byte
	Nonce     uint32 // 4 bytes
}

// BGPMssgHolder if part of the transaction that carries the actual application data in this case BGP info
type BGPMssgHolder struct {
	Source      int32   //Source of the transaction - AS number
	Destination []int32 // Array of destination nodes' AS numbers
	Reference   [32]byte
	Prefixes    []string //['192.168.0.1.24','192.168.0.1.25','192.168.0.1.26']
	StartDate   int64
	EndDate     int64
}

// ASNode represents one vertex in the AS network graph
type ASNode struct {
	ASno        int32
	PuK         [65]byte //Public key of the AS blockchain node
	IP          [4]int32
	Connections []int32 //hash of the ASNode struct of peers
}

// prefixGraph is the graph representing the network structure of the
type prefixGraph struct {
	Mux    sync.Mutex
	Graph  map[string]ASNode // ASno maps to the ASnode
	Length int
}

// ShardSignal ds is recieved from discovery to initiate sharding
type ShardSignal struct {
	Identifier [32]byte
	From       [65]byte
}

// ShardTransaction transaction to start sharding
type ShardTransaction struct {
	Identifier [32]byte
	Timestamp  int64
	From       [65]byte
	IP         [4]byte
	ShardNo    uint32
	Nonce      uint32
}

// ForwardTx ...
type ForwardTx struct {
	Tx        Transaction
	Signature []byte
	Peer      net.Conn
	Forward   bool
}

type ShardTransactionCh struct {
	Tx   ShardTransaction
	Sign []byte
}

// Peers maintains the list of all peers connected to the node
type Peers struct {
	Mux sync.Mutex
	Fds map[string]net.Conn
}

// Vertex is a wrapper struct of Transaction
type Vertex struct {
	Tx         Transaction
	Signature  []byte
	Neighbours []string
}

// DAG defines the data structure to store the blockchain
type DAG struct {
	Mux       sync.Mutex
	Genisis   string
	Graph     map[string]Vertex
	Length    int
	StorageCh chan ForwardTx
}
