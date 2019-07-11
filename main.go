package main 

import (
	dt "GO-DAG/DataTypes"
	"GO-DAG/server"
	"GO-DAG/client"
	"GO-DAG/storage"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/Discovery"
	"net"
	"fmt"
	"time"
	"math/rand"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var dag dt.DAG
	var peers dt.Peers
	dag.Graph = make(map[string] dt.Node)
	peers.Fds = make(map[string] net.Conn)
	var genisis dt.Transaction
	genisis.Timestamp = 0
	genisis.Value = 0
	s := serialize.SerializeData(genisis)
	h := Crypto.Hash(s)
	dag.Genisis = Crypto.EncodeToHex(h[:])
	var node dt.Node
	node.Tx = genisis
	dag.Graph[dag.Genisis] = node
	ch := make(chan dt.Transaction)
	go storage.AddTransaction(dag,ch)
	go server.StartServer(peers,ch)
	time.Sleep(time.Second)
	//conn,_ := net.Dial("tcp","127.0.0.1:9000")
	ips := Discovery.GetIps("ips.txt")
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	fmt.Println("connection established with all peers")
	PrivateKey := Crypto.GenerateKeys()
	client.SimulateClient(peers,PrivateKey,dag,ch)
}