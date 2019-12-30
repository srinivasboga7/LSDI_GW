package main 

import (
	"GO-DAG/server"
	"GO-DAG/client"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/Discovery"
	"GO-DAG/storage"
	"GO-DAG/sync"
	log "GO-DAG/logdump"
	"encoding/json"
	"net"
	"time"
	"os"
	"math/rand"
	"strings"
	"fmt"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var dag storage.DAG
	var peers storage.Peers
	dag.Graph = make(map[string] storage.Vertex)
	peers.Fds = make(map[string] net.Conn)
	var srv server.Server
	srv.Peers = &peers
	srv.Dag = &dag
	go srv.StartServer()
	time.Sleep(time.Second)
	log.Println("REQUESTING DISCOVERY NODE FOR PEERS")
	fmt.Println()
	ips := Discovery.GetIps("169.254.175.29:8000")
	log.Println("CONNECTING WITH PEERS")
	peers.Mux.Lock()
	peers.Fds = Discovery.ConnectToServer(ips)
	peers.Mux.Unlock()
	//time.Sleep(time.Second)
	log.Println("CONNECTION ESTABLISHED WITH ALL PEERS")
	fmt.Println()
	log.Println("STARTING TO SYNC DAG")
	fmt.Println()
	sync.copyDAG(&dag,&peers,peers.Fds[ips[0][:strings.IndexByte(ips[0],':')]])
	log.Println("DAG SYNCED")
	fmt.Println()
	// log.DefaultPrint("==========================================")
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
	log.Println("GATEWAY NODE ACTIVE")
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