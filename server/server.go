package server

import(
	"fmt"
	"net"
	dt "GO-DAG/DataTypes"
	"bytes"
	"encoding/binary"
)


func HandleConnection(connection net.Conn) {
	for {
		buf := make([]byte,1024)
		_, err := connection.Read(buf)
		if err != nil {
			// Remove from the list of the peer
			fmt.Println(err)
			break
		}
		HandleRecievedData(buf)
	}
	defer connection.Close()
}

func Deserialize(b []byte) dt.Transaction {
	r := bytes.NewReader(b)
	var tx data
	err := binary.Read(r,binary.LittleEndian,&tx)
	fmt.Println(err)
	return tx
}

func HandleRecievedData (data []byte) {
	tx := Deserialize(data)
	if ValidTransaction(tx) {
		
	}

}


func ValidTransaction(t dt.Transaction) bool {
	
	return true
	
}


func AddTransaction(t dt.Transaction) {

}

func BroadcastTransaction(t []byte, p dt.Peers, IP string) {	
	p.Mux.Lock()

	p.Mux.Unlock()
}


func StartServer() {
	listener, _ := net.Listen("tcp",":9000")
	for {
		conn, _ := listener.Accept()
		go HandleConnection(conn)

	}
	defer listener.Close()
}