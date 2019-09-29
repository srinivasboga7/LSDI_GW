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
	"os"
	"math/rand"
	"strings"
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
	copyDAG(&dag,&peers,peers.Fds[ips[0][:strings.IndexByte(ips[0],':')]])
	fmt.Println("DAG synced")
	PrivateKey := Crypto.GenerateKeys()
	var url string
	url = os.Args[1]
	var cli client.Client
	cli.Peers = &peers
	cli.Dag = &dag
	cli.PrivateKey = PrivateKey
	cli.RecieveSensorData(url)
	// client.SimulateClient(&peers,PrivateKey,&dag,url)
}


func copyDAG(dag *dt.DAG, p *dt.Peers, conn net.Conn) {
	var magicNumber uint32
	magicNumber = 2
	b := serialize.EncodeToBytes(magicNumber)
	var txs []string
	p.Mux.Lock()
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
	p.Mux.Unlock()
	fmt.Println(len(txs))
	magicNumber = 3
	num := serialize.EncodeToBytes(magicNumber)
	var v string
	for _,v = range txs {
		hash := Crypto.DecodeToBytes(v)
		hash = append(num,hash...)
		conn.Write(hash)
		buf := make([]byte,1024)
		l,_ := conn.Read(buf)
		tx,sign := serialize.DeserializeTransaction(buf[:l])
		// storage.AddTransaction(dag,tx,sign)
		var vertex dt.Vertex
		vertex.Tx = tx
		vertex.Signature = sign
		var tip [32]byte 
		dag.Mux.Lock()
		if tx.LeftTip == tip && tx.RightTip == tip {
			dag.Genisis = v
			fmt.Println("Genisis Transaction")
		}
		dag.Graph[v] = vertex
		dag.Mux.Unlock()
	}
	constructDAG(dag)
}

func constructDAG(dag *dt.DAG) {
	for h,vertex := range dag.Graph {
		tx := vertex.Tx
		left := Crypto.EncodeToHex(tx.LeftTip[:])
		right := Crypto.EncodeToHex(tx.RightTip[:])
		if left != right {
			if _,ok := dag.Graph[left] ; ok {
				l := dag.Graph[left]
				l.Neighbours = append(l.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			}
			if _,ok := dag.Graph[right] ; ok {
				r := dag.Graph[right]
				r.Neighbours = append(r.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])] = r
			}
		} else {
			if _,ok := dag.Graph[left] ; ok {
				l := dag.Graph[left]
				l.Neighbours = append(l.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			}
		}
	}

	// sync with the orphaned Transactions
	for _,vertices := range storage.OrphanedTransactions {
		for _, v := range vertices {
			storage.AddTransaction(dag,v.Tx,v.Signature)
		}
	}
}
