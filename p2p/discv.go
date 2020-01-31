package p2p

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// FindPeers fetches some peers by querying the discovery service
// The Bootstrap disocvery nodes are provided in a file
func FindPeers(host *PeerID) []PeerID {
	var peers []PeerID

	// reading from the file containing discv nodes
	f, err := os.Open("bootstrapNodes.txt")
	if err != nil {
		log.Fatal("Problem opening the file containing bootstrap nodes")
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("Incosistent file containing bootstrap nodes")
	}

	// bootstrapNodes is the list of some of the discv nodes
	bootstrapNodes := strings.Split(string(b), "\n")

	// iteratively trying to query discovery nodes until a successful response
	for _, addr := range bootstrapNodes {
		peers, err = queryDiscoveryService(addr, host)
		if err != nil {
			log.Println("Failed to query discv node", addr)
		} else {
			break
		}
	}

	return peers
}

func queryDiscoveryService(servAddr string, localID *PeerID) ([]PeerID, error) {

	// querying the discovery service
	conn, err := net.Dial("tcp", servAddr)
	log.Println("Dialed the discovery server")
	if err != nil {
		return nil, err
	}
	ip := conn.LocalAddr().String()
	ip = ip[:strings.IndexByte(ip, ':')]
	localID.IP = serializeIPAddr(ip)
	query := []byte{0x04}
	query = append(query, localID.IP...)
	query = append(query, localID.PublicKey...)
	_, err = conn.Write(query)
	if err != nil {
		log.Println(err)
	}
	buf := make([]byte, 1024)

	// response from discovery service
	l, _ := conn.Read(buf)
	var peers [][]byte
	err = json.Unmarshal(buf[:l], &peers)
	if err != nil {
		return nil, err
	}
	var p []PeerID
	for _, peer := range peers {
		var s PeerID
		s.IP = peer[:4]
		s.PublicKey = peer[4:]
		p = append(p, s)
	}
	return p, nil
}

func serializeIPAddr(IP string) []byte {
	ipFields := strings.Split(IP, ".")
	buf := new(bytes.Buffer)
	var i int
	for _, fields := range ipFields {
		i, _ = strconv.Atoi(fields)
		binary.Write(buf, binary.LittleEndian, uint8(i))
	}
	return buf.Bytes()
}
