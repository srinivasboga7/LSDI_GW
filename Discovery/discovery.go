package Discovery

import (
	//"io/ioutil"
	//"strings"
	"bytes"
	"log"
	"net"
	"strconv"

	//dt "GO-DAG/DataTypes"
	//"log"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Request struct {
	// NumberOfPeers int
	NodeType string
}

type peerAddr struct {
	IP        []byte
	Publickey []byte
}

func GetIps(Addrs string) []string {
	//Returns slice which has IPs as strings
	//Currently getting from file
	//Later get from DNS server or anything else
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", Addrs)
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	var req Request
	req.NodeType = "GatewayNode"
	request, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}
	conn.Write(request)
	buf := make([]byte, 1024)
	l, _ := conn.Read(buf)
	var IPList []string
	err = json.Unmarshal(buf[:l], &IPList)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(IPList)
	return IPList
}

func ConnectToServer(ips []string) map[string]net.Conn {
	//Takes a map of ips to socket file descriptors of tcp connections
	p := make(map[string]net.Conn)
	for _, ip := range ips {
		tcpAddr, _ := net.ResolveTCPAddr("tcp4", ip)
		for {
			conn, err := net.DialTCP("tcp", nil, tcpAddr) //BLocking call
			if err == nil {
				ip = ip[:strings.IndexByte(ip, ':')]
				p[ip] = conn
				break
			}
			time.Sleep(time.Second)
		}
	}
	return p
}

func queryDiscoveryService(localAddr string, pubKey []byte, servAddr string) []peerAddr {
	// serializing the IP address and constructing the query
	ip := strings.Split(localAddr, ".")
	query := []byte{0x04}
	for _, n := range ip {
		buf := new(bytes.Buffer)
		num, _ := strconv.Atoi(n)
		binary.Write(buf, binary.LittleEndian, uint8(num))
		query = append(query, buf.Bytes()...)
	}
	query = append(query, pubKey...)

	// querying the discovery service

	udpAddr, _ := net.ResolveUDPAddr("udp4", servAddr)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal("Discovery Service Down")
	}
	conn.WriteToUDP(query, udpAddr)
	buf := make([]byte, 1024)

	// response from discovery service
	l, _, _ := conn.ReadFromUDP(buf)
	var peers []peerAddr
	json.Unmarshal(buf[:l], &peers)

	// sending the acknowledgment
	conn.WriteToUDP([]byte{0x05}, udpAddr)

	return peers
}

func udpServer() {
	addr, _ := net.ResolveUDPAddr("udp4", ":8080")
	for {
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			continue
		}
		buf := make([]byte, 1)
		_, dest, _ := conn.ReadFromUDP(buf)
		if buf[0] == 0x01 {
			conn.WriteToUDP([]byte{0x02}, dest)
		}
		conn.Close()
	}
}
