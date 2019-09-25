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
	"os"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var dag dt.DAG
	var peers dt.Peers
	dag.Graph = make(map[string] dt.Vertex)
	peers.Fds = make(map[string] net.Conn)
	var genesis dt.Transaction
	copy(genesis.TxID[:],[]byte("1234567812345678"))
	s := serialize.SerializeData(genesis)
	h := Crypto.Hash(s)
	dag.Genisis = Crypto.EncodeToHex(h[:])
	var node dt.Vertex
	node.Tx = genesis
	dag.Graph[dag.Genisis] = node
	var srv server.Server
	srv.Peers = &peers
	srv.Dag = &dag
	go srv.StartServer()
	time.Sleep(time.Second)
	ips := Discovery.GetIps("192.168.122.190:8000")
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	fmt.Println("connection established with all peers")
	PrivateKey := Crypto.GenerateKeys()
	var url string
	url = os.Args[1]
	var cli client.Client
	cli.Peers = &peers
	cli.Dag = &dag
	cli.PrivateKey = PrivateKey
	cli.RecieveSensorData(url)
}