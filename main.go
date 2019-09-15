package main 

import (
	dt "GO-DAG/DataTypes"
	"GO-DAG/server"
	"GO-DAG/client"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/Discovery"
	"GO-DAG/storage"
	"encoding/json"
	"net"
	"fmt"
	"time"
	"math/rand"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var dag dt.DAG
	var peers dt.Peers
	dag.Graph = make(map[string] dt.Vertex)
	peers.Fds = make(map[string] net.Conn)
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
	time.Sleep(time.Second)
	CopyDAG(&dag,&peers)
	fmt.Println("DAG synced")
	PrivateKey := Crypto.GenerateKeys()
	client.SimulateClient(&peers,PrivateKey,&dag)
}


func CopyDAG(dag *dt.DAG, p *dt.Peers) {
	// copy only the tips of the DAG.
	var magic_number uint32
	magic_number = 2
	b := serialize.EncodeToBytes(magic_number)
	var conn net.Conn
	var txs []string
	p.Mux.Lock()
	for _,conn = range p.Fds {
		conn.Write(b)
		var ser []byte
		for { 
			buf := make([]byte,1024)
			l,_ := conn.Read(buf)
			ser = append(ser,buf[:l]...)
			if l < 1024 {
				break
			}
		}
		json.Unmarshal(ser,&txs)
		if len(txs) != 0 {
			break
		}
	}
	p.Mux.Unlock()
	fmt.Println(len(txs))
	magic_number = 3
	num := serialize.EncodeToBytes(magic_number)
	var v string
	for _,v = range txs {
		hash := Crypto.DecodeToBytes(v)
		hash = append(num,hash...)
		conn.Write(hash)
		buf := make([]byte,1024)
		l,_ := conn.Read(buf)
		tx,sign := serialize.DeserializeTransaction(buf[:l])
		var node dt.Vertex
		node.Tx = tx
		node.Signature = sign
		dag.Mux.Lock()
		dag.Graph[v] = node
		dag.Mux.Unlock() 
	}
	dag.Mux.Lock()
	dag.Genisis = v
	dag.Mux.Unlock()

	// sync with the already arrived transactions
	for _,node := range storage.OrphanedTransactions {
		storage.AddTransaction(dag,node.Tx,node.Signature)
	}
}