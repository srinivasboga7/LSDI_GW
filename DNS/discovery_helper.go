package main

import (
	"net"
	"log"
	"sync"
	"encoding/json"
	d "GO-DAG/Discovery"
	"math/rand"
	"strings"
	"fmt"
)

type ActiveNodes struct {
	Mux sync.Mutex 
	GatewayNodeAddrs []string
	StorageNodeAddrs []string
}

func main() {
	var nodes ActiveNodes
	Listener,err := net.Listen("tcp",":8000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn,_ := Listener.Accept()
		go HandleRequest(conn,&nodes)
	}
}

func HandleRequest(conn net.Conn, nodes *ActiveNodes) {
	buf := make([]byte,1024)
	l,err := conn.Read(buf)
	if err != nil {
		return
	}
	addr := conn.RemoteAddr().String()
	ip := addr[:strings.IndexByte(addr,':')]
	ip = ip + ":9000" 
	var req d.Request
	err = json.Unmarshal(buf[:l],&req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(req.NodeType)
	var randomnodes []string 
	if req.NodeType == "StorageNode" {
		return
	} else if req.NodeType == "GatewayNode"{
		nodes.Mux.Lock()
		ActiveNodes := len(nodes.GatewayNodeAddrs)
		var perm []int
		if ActiveNodes != 0 {
			perm = rand.Perm(ActiveNodes)
			if ActiveNodes > 1 {
				randomnodes = append(randomnodes,nodes.GatewayNodeAddrs[perm[0]])
				randomnodes = append(randomnodes,nodes.GatewayNodeAddrs[perm[1]])
			} else {
				randomnodes = append(randomnodes,nodes.GatewayNodeAddrs[perm[0]])
			}
		}
		reply,_ := json.Marshal(randomnodes)
		conn.Write(reply)
		nodes.GatewayNodeAddrs = append(nodes.GatewayNodeAddrs,ip)
		nodes.Mux.Unlock()
		conn.Close()
	} else {
		return
	}
	defer conn.Close()
}