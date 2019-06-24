package server

import(
	"fmt"
	"net"
	"encoding/json"
	"GO-DAG/DataTypes"
)


func HandleConnection(Connection net.Conn,IPs []string) {

}


func ValidTransaction(t DataTypes.Transaction) bool {

	return true
	
}

/*
func AddTransaction(t DataTypes.Transaction) {

}
*/

func BroadcastTransaction(t []bytes, IPs []string) {


}


func StartServer(NodeIPs []string) {

	for {

		listener, _ := net.Listen("tcp",":9000")
		conn, _ := listener.Accept()
		go HandleConnection(conn,NodeIPs)

	}
	defer listener.Close()
}