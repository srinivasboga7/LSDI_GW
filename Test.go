package main

import(
	"GO-DAG/server"
	"GO-DAG/client"
	dt "GO-DAG/DataTypes"
	"net"
	"time"
)


func main() {
	go server.StartServer()
	time.Sleep(time.Second)
	var p dt.Peers
	p.Fds = make([]net.Conn,1)
	conn,_ := net.Dial("tcp","127.0.0.1:9000")
	p.Fds[0] = conn
	client.SimulateClient(p)
}