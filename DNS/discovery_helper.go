package main

import (
	"time"
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
	rand.Seed(time.Now().UnixNano())
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
		nodes.Mux.Lock()
		ActiveNodes_storage := len(nodes.StorageNodeAddrs)
		ActiveNodes_gateway := len(nodes.GatewayNodeAddrs)
		var perm1 []int
		var perm2 []int
		if ActiveNodes_storage != 0 {
			if ActiveNodes_gateway != 0 {
				perm1 = rand.Perm(ActiveNodes_storage)
				perm2 = rand.Perm(ActiveNodes_gateway)
				randomnodes = append(randomnodes,nodes.StorageNodeAddrs[perm1[0]])
				randomnodes = append(randomnodes,nodes.GatewayNodeAddrs[perm2[0]])
			} else {
				perm1 = rand.Perm(ActiveNodes_storage)
				if ActiveNodes_storage > 1 {
					randomnodes = append(randomnodes,nodes.StorageNodeAddrs[perm1[0]])
					randomnodes = append(randomnodes,nodes.StorageNodeAddrs[perm1[1]])
				} else {
					randomnodes = append(randomnodes,nodes.StorageNodeAddrs[perm1[0]])
				}
			}
		}
		reply,_ := json.Marshal(randomnodes)
		conn.Write(reply)
		nodes.StorageNodeAddrs = append(nodes.StorageNodeAddrs,ip)
		fmt.Println(nodes.StorageNodeAddrs)
		fmt.Println(nodes.GatewayNodeAddrs)
		nodes.Mux.Unlock()
		conn.Close()
	} else if req.NodeType == "GatewayNode"{
		nodes.Mux.Lock()
		AllNodes := append(nodes.GatewayNodeAddrs,nodes.StorageNodeAddrs...)
		ActiveNodes := len(AllNodes)
		ActiveNodes_storage := len(nodes.StorageNodeAddrs)
		ActiveNodes_gateway := len(nodes.GatewayNodeAddrs)
		var perm1 []int
		var perm2 []int
		if ActiveNodes_gateway == 0 {
			perm1 = rand.Perm(ActiveNodes_storage)
			randomnodes = append(randomnodes,nodes.StorageNodeAddrs[perm1[0]])
			if ActiveNodes_storage > 1 {
				randomnodes = append(randomnodes,nodes.StorageNodeAddrs[perm1[1]])
			}
		} else if ActiveNodes_gateway == 1 {
			perm1 = rand.Perm(ActiveNodes_storage)
			perm2 = rand.Perm(ActiveNodes_gateway)
			randomnodes = append(randomnodes,nodes.GatewayNodeAddrs[perm2[0]])
			randomnodes = append(randomnodes,nodes.StorageNodeAddrs[perm1[0]])
		} else {
			perm1 = rand.Perm(ActiveNodes)
			perm2 = rand.Perm(ActiveNodes_gateway)
			randomnodes = append(randomnodes,nodes.GatewayNodeAddrs[perm2[0]])
			if randomnodes[0] == AllNodes[perm1[0]] {
				randomnodes = append(randomnodes,AllNodes[perm1[1]])
			} else {
				randomnodes = append(randomnodes,AllNodes[perm1[0]])
			}
		}
		reply,_ := json.Marshal(randomnodes)
		conn.Write(reply)
		nodes.GatewayNodeAddrs = append(nodes.GatewayNodeAddrs,ip)
		fmt.Println(nodes.StorageNodeAddrs)
		fmt.Println(nodes.GatewayNodeAddrs)
		nodes.Mux.Unlock()
		conn.Close()
	} else if req.NodeType == "UserNode"{
		nodes.Mux.Lock()
		ActiveNodes_storage := len(nodes.StorageNodeAddrs)
		var perm1 []int
		perm1 = rand.Perm(ActiveNodes_storage)
		var randomNode string
		randomNode = nodes.StorageNodeAddrs[perm1[0]]
		nodes.Mux.Unlock()
		reply := []byte(randomNode)
		conn.Write(reply)
		conn.Close()
	} else {
		return
	}
	defer conn.Close()
}