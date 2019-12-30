package p2p

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

// FindPeers fetches some peers by querying the discovery service
// The Bootstrap disocvery nodes are provided in a file
func FindPeers(host PeerID) []PeerID {
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

	// check if the query to all discovery nodes are unsuccessful
	if len(peers) == 0 {
		log.Fatal("Discovery service down")
	}

	return peers
}

func queryDiscoveryService(servAddr string, localID PeerID) ([]PeerID, error) {

	query := []byte{0x04}
	query = append(query, localID.IP...)
	query = append(query, localID.PublicKey...)

	// querying the discovery service

	udpAddr, _ := net.ResolveUDPAddr("udp4", servAddr)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	conn.WriteToUDP(query, udpAddr)
	buf := make([]byte, 1024)

	// response from discovery service
	l, _, _ := conn.ReadFromUDP(buf)
	var peers []PeerID
	err = json.Unmarshal(buf[:l], &peers)
	if err != nil {
		return nil, err
	}

	// sending the acknowledgment
	conn.WriteToUDP([]byte{0x05}, udpAddr)

	return peers, nil
}
