package server

import(
	"fmt"
	"net"
	dt "GO-DAG/DataTypes"
	"bytes"
	"encoding/gob"
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

func HandleRecievedData(data []byte) {
	var t dt.Transaction
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&t)
	fmt.Println(t)
}

func Deserialize (data []byte) {

}


func ValidTransaction(t DataTypes.Transaction) bool {

	return true
	
}


func AddTransaction(t DataTypes.Transaction) {

}

func BroadcastTransaction(t []byte, p DataTypes.peers) {	
	p.mux.Lock()

	p.mux.UnLock()
}


func StartServer() {
	listener, _ := net.Listen("tcp",":9000")
	for {
		conn, _ := listener.Accept()
		go HandleConnection(conn)

	}
	defer listener.Close()
}