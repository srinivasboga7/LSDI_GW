package DataTypes


import(
	"sync"
	"net"
)


type Transaction struct {
	// Definition of data
	Timestamp int64
	Value float64
	//from string
	//signature string
	//LeftTip string
	//RightTip string
	//nonce uint32
}

type Peers struct {
	Mux sync.Mutex
	Fds []net.Conn
}

type Store struct {
	Mux sync.Mutex
	DAG map[string] dt.Transaction
}