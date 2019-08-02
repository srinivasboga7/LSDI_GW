package main 

import (
	dt "GO-DAG/DataTypes"
	"GO-DAG/server"
	"GO-DAG/client"
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
	go server.StartServer(peers,dag)
	time.Sleep(time.Second)
	ips := Discovery.GetIps("ips.txt")
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	fmt.Println("connection established with all peers")
	PrivateKey := Crypto.GenerateKeys()
	client.SimulateClient(peers,PrivateKey,dag)
}