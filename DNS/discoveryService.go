package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
)

const (
	maxNodes = 10
)

type peerAddr struct {
	ShardID     []byte
	networkAddr []byte
	PublicKey   []byte
}

// better add a database as the backup
type liveNodes struct {
	GWNodes gatewayNodes
	SNNodes storageNodes
}

// stores all the gateway nodes in a shard
type shardNodes struct {
	ShardID []byte
	Nodes   []peerAddr
}

type gatewayNodes struct {
	mux    sync.RWMutex
	shards []shardNodes
}

type storageNodes struct {
	mux   sync.RWMutex
	Nodes []peerAddr
}

func find(list []peerAddr, element peerAddr) bool {
	for _, node := range list {
		if bytes.Compare(element.networkAddr, node.networkAddr) == 0 {
			return true
		}
	}
	return false
}

func (nodes *liveNodes) appendTo(peer peerAddr, GS bool) {
	if GS {
		nodes.GWNodes.mux.Lock()
		for _, shard := range nodes.GWNodes.shards {
			if bytes.Compare(peer.ShardID, shard.ShardID) == 0 {
				shard.Nodes = append(shard.Nodes, peer)
				if len(shard.Nodes) >= maxNodes {
					go nodes.initiateSharding(shard)
				}
				break
			}
		}
		nodes.GWNodes.mux.Unlock()
	} else {
		nodes.SNNodes.mux.Lock()
		nodes.SNNodes.Nodes = append(nodes.SNNodes.Nodes, peer)
		nodes.SNNodes.mux.Unlock()
	}
}

func (nodes *liveNodes) remove(GWNodes []peerAddr, SNNodes []peerAddr) {
	// temporarily halt , write after testing the other shit
}

func constructShardSignal() []byte {
	shardSignal := []byte{0x35}
	shardSignal = append(shardSignal, []byte{0x00}...)
	return shardSignal
}

func constructPing() []byte {
	ping := []byte{0x01}
	ping = append(ping, []byte{0x00}...)
	return ping
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

func sendShardSignal(nodes shardNodes) {
	randNode := nodes.Nodes[rand.Intn(len(nodes.Nodes))]
	conn, _ := net.Dial("tcp", parseAddr(randNode.networkAddr))
	b := constructShardSignal()
	conn.Write(b)
	conn.Close()
}

func (nodes *liveNodes) updateShard(ShardID []byte, IP []byte, PubKey []byte) {
	nodes.GWNodes.mux.Lock()
	var peer peerAddr
	peer.networkAddr = IP
	peer.PublicKey = PubKey
	peer.ShardID = ShardID
	for _, shard := range nodes.GWNodes.shards {
		if bytes.Compare(ShardID, shard.ShardID) == 0 {
			shard.Nodes = append(shard.Nodes, peer)
			break
		}
	}
	nodes.GWNodes.mux.Unlock()
}

func (nodes *liveNodes) initiateSharding(prevShard shardNodes) {
	sendShardSignal(prevShard)
	nodes.GWNodes.mux.Lock()
	for i, shard := range nodes.GWNodes.shards {
		if bytes.Compare(prevShard.ShardID, shard.ShardID) == 0 {
			// deleting the old shard
			nodes.GWNodes.shards[i] = nodes.GWNodes.shards[len(nodes.GWNodes.shards)-1]
			nodes.GWNodes.shards = nodes.GWNodes.shards[:len(nodes.GWNodes.shards)-1]
			break
		}
	}
	// creating two new shards
	var nextShard1, nextShard2 shardNodes
	nextShard1.ShardID = append(prevShard.ShardID, 0x00)
	nextShard2.ShardID = append(prevShard.ShardID, 0x01)
	nodes.GWNodes.shards = append(nodes.GWNodes.shards, nextShard1, nextShard2)
	nodes.GWNodes.mux.Unlock()
}

// temporary method optimize if needed
func getRandomPeers(nodes *liveNodes, peer peerAddr, gs bool) []peerAddr {
	// find an efficient way to select the peers
	var peers []peerAddr
	if gs {
		nodes.GWNodes.mux.RLock()
		nodes.SNNodes.mux.RLock()
		var shard shardNodes
		for _, shard = range nodes.GWNodes.shards {
			if bytes.Compare(shard.ShardID, peer.ShardID) == 0 {
				break
			}
		}
		if len(shard.Nodes) > 0 {
			perm1 := rand.Perm(len(shard.Nodes))
			perm2 := rand.Perm(len(nodes.SNNodes.Nodes))
			peers = append(peers, shard.Nodes[perm1[0]])

			if rand.Intn(2) == 0 {
				peers = append(peers, shard.Nodes[perm1[1]])
			} else {
				peers = append(peers, nodes.SNNodes.Nodes[perm2[0]])
			}
		}
		nodes.SNNodes.mux.RUnlock()
		nodes.GWNodes.mux.RUnlock()
	} else {
		nodes.SNNodes.mux.RLock()
		l := len(nodes.SNNodes.Nodes)
		perm := rand.Perm(l)
		if l == 1 {
			peers = append(peers, nodes.SNNodes.Nodes[perm[0]])
		} else if l > 1 {
			peers = append(peers, nodes.SNNodes.Nodes[perm[0]])
			peers = append(peers, nodes.SNNodes.Nodes[perm[1]])
		}
		nodes.SNNodes.mux.RUnlock()
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
		log.Println(err)
		return
	}
	if buffer[0] == 0x06 {
		IP, PubKey, ShardID := buffer[1:5], buffer[5:70], buffer[70:l]
		nodes.updateShard(ShardID, IP, PubKey)
	} else {
		IP, PubKey, GS := deserialize(buffer[:l])
		var newPeer peerAddr
		if GS {
			x := rand.Intn(len(nodes.GWNodes.shards))
			newPeer.ShardID = nodes.GWNodes.shards[x].ShardID
		} else {
			newPeer.ShardID = []byte{0x2}
		}
		newPeer.PublicKey = PubKey
		newPeer.networkAddr = IP
		randomPeers := getRandomPeers(nodes, newPeer, GS)
		var p [][]byte
		for _, peer := range randomPeers {
			p = append(p, append(append(peer.networkAddr, peer.PublicKey...), peer.ShardID...))
		}
		b, _ := json.Marshal(p)
		conn.Write(b)
		nodes.appendTo(newPeer, GS)
	}
	return
}

// add a database to store the peers
func checkDB(nodes *liveNodes) {
	//
	return
}

func main() {
	port := os.Args[1]
	log.Println("Discovery service running on port", port)
	var nodes liveNodes
	var firstShard shardNodes
	firstShard.ShardID = []byte{0x00}
	nodes.GWNodes.shards = append(nodes.GWNodes.shards, firstShard)
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
		log.Println("New connection", conn.RemoteAddr().String())
		go handleConnection(conn, &nodes)
	}
}
