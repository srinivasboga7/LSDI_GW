package main 

import (
	dt "GO-DAG/DataTypes"
	"GO-DAG/server"
	"GO-DAG/client"
	"GO-DAG/storage"
	"GO-DAG/Crypto"
	"net"
	//"fmt"
	"time"
)

func main() {
	var dag dt.DAG
	var peers dt.Peers
	dag.Graph = make(map[string] dt.Node)
	peers.Fds = make(map[string] net.Conn)
	ch := make(chan dt.Transaction)
	go storage.AddTransaction(dag,ch)
	go server.StartServer(peers,ch)
	time.Sleep(time.Second)
	conn,_ := net.Dial("tcp","127.0.0.1:9000")
	peers.Mux.Lock()
	peers.Fds["127.0.0.1"] = conn
	peers.Mux.Unlock()
	PrivateKey := Crypto.LoadKeys()
	client.SimulateClient(peers,PrivateKey)
}