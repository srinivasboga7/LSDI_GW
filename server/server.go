package server

import(
	"fmt"
	"net"
	dt "GO-DAG/DataTypes"
	"bytes"
	"encoding/binary"
	"strings"
)

// still have to add DAG

func HandleConnection(connection net.Conn, p dt.Peers ) {
	// each connection is handled in a seperate go routine
	for {
		buf := make([]byte,1024)
		_, err := connection.Read(buf)
		if err != nil {
			// Remove from the list of the peer
			fmt.Println(err)
			break
		}
		addr := conn.RemoteAddr()
		ip := addr[:strings.IndexByte(addr,':')] // seperate the port to get only IP
		HandleRecievedData(buf,ip,p)
	}
	defer connection.Close()
}

func Deserialize(b []byte) dt.Transaction {
	// only a temporary method will change to include signature and other checks
	r := bytes.NewReader(b)
	var tx data
	err := binary.Read(r,binary.LittleEndian,&tx)
	if err != nil { 
		fmt.Println(err)
	}
	return tx
}

func HandleRecievedData (data []byte, IP string, p dt.Peers) {
	tx := Deserialize(data)
	if ValidTransaction(tx) {
		ForwardTransaction(data,IP,p)
		AddTransaction(tx)
	}
}


func ValidTransaction(t dt.Transaction, signature []byte) bool {
	// check the signature 
	return true
}



func AddTransaction(t dt.Transaction) {
	// Add the transaction to the DAG

}

func ForwardTransaction(t []byte, IP string, p dt.Peers) {
	// sending the transaction to the peers excluding the one it came from	
	p.Mux.Lock()
	for k,conn := range p {
		if k != IP {
			conn.Write(t)
		} 
	}
	p.Mux.Unlock()
}


func StartServer(p dt.Peers) {
	listener, _ := net.Listen("tcp",":9000")
	for {
		conn, _ := listener.Accept()
 		go HandleConnection(conn,p) // go routine executes concurrently

	}
	defer listener.Close()
}