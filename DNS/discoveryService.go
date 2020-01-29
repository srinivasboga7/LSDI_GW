package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type peerAddr struct {
	GWSN        bool
	networkAddr []byte
	PublicKey   []byte
}

// better add a database as the backup
type liveNodes struct {
	mux    sync.RWMutex
	GWAddr []peerAddr
	SNAddr []peerAddr
}

func find(list []peerAddr, element peerAddr) bool {
	for _, node := range list {
		if bytes.Compare(element.networkAddr, node.networkAddr) == 0 {
			return true
		}
	}
	return false
}

func (nodes *liveNodes) appendTo(IP []byte, PublicKey []byte, GS bool) {
	var newPeer peerAddr
	newPeer.networkAddr = IP
	newPeer.PublicKey = PublicKey
	nodes.mux.Lock()
	if GS {
		if find(nodes.GWAddr, newPeer) {
			fmt.Println("Active Node requesting the peers")
		} else {
			nodes.GWAddr = append(nodes.GWAddr, newPeer)
		}
	} else {
		if find(nodes.SNAddr, newPeer) {
			fmt.Println("Active Node requesting the peers")
		} else {
			nodes.SNAddr = append(nodes.SNAddr, newPeer)
		}
	}
	nodes.mux.Unlock()
}

func (nodes *liveNodes) remove(GWNodes []peerAddr, SNNodes []peerAddr) {
	nodes.mux.Lock()
	for _, gwnode := range GWNodes {
		for i, node := range nodes.GWAddr {
			if bytes.Compare(gwnode.networkAddr, node.networkAddr) == 0 {
				nodes.GWAddr[i] = nodes.GWAddr[len(nodes.GWAddr)-1]
				nodes.GWAddr = nodes.GWAddr[:len(nodes.GWAddr)-1]
			}
		}
	}
	for _, snnode := range SNNodes {
		for i, node := range nodes.SNAddr {
			if bytes.Compare(snnode.networkAddr, node.networkAddr) == 0 {
				nodes.SNAddr[i] = nodes.SNAddr[len(nodes.SNAddr)-1]
				nodes.SNAddr = nodes.SNAddr[:len(nodes.SNAddr)-1]
			}
		}
	}
	nodes.mux.Unlock()
}

func constructPing() []byte {
	return []byte{0x01}
}

func checkPong(b []byte) bool {
	if b[0] == 0x02 {
		return true
	}
	return false
}

func parseAddr(b []byte) string {
	var IP struct {
		b1 uint8
		b2 uint8
		b3 uint8
		b4 uint8
	}
	buf := bytes.NewReader(b)
	binary.Read(buf, binary.LittleEndian, &IP)
	addr := strconv.Itoa(int(IP.b1)) + "." + strconv.Itoa(int(IP.b2)) + "."
	addr += strconv.Itoa(int(IP.b3)) + "." + strconv.Itoa(int(IP.b4)) + ":8080"
	return addr
}

func pingNodes(nodes *liveNodes, interval int) {
	var expiredGWNodes []peerAddr
	var GWNodes []peerAddr
	var SNNodes []peerAddr

	nodes.mux.RLock()
	copy(GWNodes, nodes.GWAddr)
	copy(SNNodes, nodes.SNAddr)
	nodes.mux.RUnlock()

	for _, node := range GWNodes {
		time.Sleep(time.Duration(interval) * time.Second)
		peerUDPAddr, _ := net.ResolveUDPAddr("udp4", parseAddr(node.networkAddr))
		conn, err := net.DialUDP("udp", nil, peerUDPAddr)
		if err != nil {
			// remove the peer
			expiredGWNodes = append(expiredGWNodes, node)
			fmt.Println(err)
			continue
		}
		conn.WriteToUDP(constructPing(), peerUDPAddr)
		buffer := make([]byte, 1)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		conn.ReadFromUDP(buffer)
		if !checkPong(buffer) {
			// remove the peer
			expiredGWNodes = append(expiredGWNodes, node)
		}
	}
	var expiredSNNodes []peerAddr
	for _, node := range SNNodes {
		time.Sleep(time.Second)
		peerUDPAddr, _ := net.ResolveUDPAddr("udp4", parseAddr(node.networkAddr))
		conn, err := net.DialUDP("udp", nil, peerUDPAddr)
		if err != nil {
			// remove the peer
			expiredSNNodes = append(expiredSNNodes, node)
			fmt.Println("Node down ", conn.RemoteAddr().String())
			fmt.Println(err)
			continue
		}
		conn.WriteToUDP(constructPing(), peerUDPAddr)
		buffer := make([]byte, 1)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		conn.ReadFromUDP(buffer)
		if !checkPong(buffer) {
			// remove the peer
			expiredSNNodes = append(expiredSNNodes, node)
			fmt.Println("Node down ", conn.RemoteAddr().String())
		}
	}
	nodes.remove(expiredGWNodes, expiredSNNodes)
}

func getRandomPeers(GWNodes []peerAddr, SNNodes []peerAddr, gs bool) []peerAddr {
	// find an efficient way to select the peers
	var peers []peerAddr
	lenGWNodes := len(GWNodes)
	lenSNNodes := len(SNNodes)
	if gs {
		if lenGWNodes == 1 {
			peers = append(peers, GWNodes[0])
			peers = append(peers, SNNodes[rand.Intn(lenSNNodes)])
		} else if lenGWNodes > 1 {
			peers = append(peers, GWNodes[rand.Intn(lenGWNodes)])
			//
		}
	} else {
		if lenSNNodes == 1 {
			peers = append(peers, SNNodes[0])
			peers = append(peers, GWNodes[rand.Intn(lenGWNodes)])
		} else if lenSNNodes > 1 {
			peers = append(peers, SNNodes[rand.Intn(lenSNNodes)])
		}
	}
	return peers
}

func deserialize(b []byte) ([]byte, []byte, bool) {
	var gs bool
	gs = false
	if b[0] == 0x04 {
		gs = true
	}
	return b[1:5], b[5:], gs
}

func handleConnection(conn net.Conn, nodes *liveNodes) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	l, err := conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	IP, PubKey, GS := deserialize(buffer[:l])
	nodes.mux.RLock()
	randomPeers := getRandomPeers(nodes.GWAddr, nodes.SNAddr, GS)
	nodes.mux.RUnlock()
	b, _ := json.Marshal(randomPeers)
	conn.Write(b)
	nodes.appendTo(IP, PubKey, GS)
	return
}

// add a database to store the peers
func checkDB(nodes *liveNodes) {
	//
	return
}

func main() {
	port := os.Args[1]
	fmt.Println("Discovery service running")
	var nodes liveNodes
	checkDB(&nodes)
	go pingNodes(&nodes, 100)

	listener, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn, &nodes)
	}
}
