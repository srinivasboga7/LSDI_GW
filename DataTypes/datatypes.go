package DataTypes


import(
	"sync"
	"net"
)


type Transaction struct {
	// Definition of data
	Timestamp int64
	Value float64
	from []byte
	LeftTip [32]byte
	RightTip [32]byte
	nonce uint32
}

type Peers struct {
	Mux sync.Mutex
	Fds map[string] net.Conn
}

type Store struct {
	Mux sync.Mutex
	DAG map[string] Transaction
}