package main 

import (
	dt "GO-DAG/DataTypes"
	"GO-DAG/server"
	"GO-DAG/client"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/Discovery"
	log "GO-DAG/logdump"
	"net"
	"time"
	"math/rand"
	"os"
	"fmt"
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
	log.Println("REQUESTING DISCOVERY NODE FOR PEERS")
	fmt.Println()
	ips := Discovery.GetIps("169.254.175.29:8000")
	log.Println("CONNECTING WITH PEERS")
	fmt.Println()
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	log.Println("CONNECTION ESTABLISHED WITH ALL PEERS")
	fmt.Println()
	var url string
	url = os.Args[1]
	/*
	var cli client.Client
	cli.Peers = &peers
	cli.Dag = &dag
	if !Crypto.CheckForKeys() {
		cli.PrivateKey = Crypto.GenerateKeys()
	} else {
		cli.PrivateKey = Crypto.LoadKeys()
	}
	//log.Println("GATEWAY NODE ACTIVE")
	fmt.Println()
	cli.RecieveSensorData(url)
	*/
	var PrivateKey Crypto.PrivateKey
	if !Crypto.CheckForKeys() {
		PrivateKey = Crypto.GenerateKeys()
	} else {
		PrivateKey = Crypto.LoadKeys()
	}

	client.SimulateClient(&peers,PrivateKey,&dag,url)
}