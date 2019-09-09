package Discovery

import (
	//"io/ioutil"
	//"strings"
	"net"
	//dt "GO-DAG/DataTypes"
	//"log"
	"fmt"
	"time"
	"encoding/json"
	"strings"
)

type Request struct {
	// NumberOfPeers int
	NodeType string
}

func GetIps(Addrs string) []string {
	//Returns slice which has IPs as strings
	//Currently getting from file
	//Later get from DNS server or anything else
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", Addrs)
	conn,_ := net.DialTCP("tcp",nil,tcpAddr)
	var req Request
	req.NodeType = "GatewayNode"
	request,err := json.Marshal(req)
	if err != nil {
		fmt.Println(err)
	}
	conn.Write(request)
	buf := make([]byte,1024)
	l,_ := conn.Read(buf)
	var IPList []string
	err = json.Unmarshal(buf[:l],&IPList)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(IPList)
	return IPList
}

func ConnectToServer(ips []string) (map[string] net.Conn) {
	//Takes a map of ips to socket file descriptors of tcp connections
	p := make(map[string] net.Conn)
	for _,ip := range ips {
		tcpAddr, _ := net.ResolveTCPAddr("tcp4", ip)
		for {
			conn,err := net.DialTCP("tcp", nil, tcpAddr)    //BLocking call
			if err == nil {
				ip = ip[:strings.IndexByte(ip,':')]
				p[ip] = conn
				break
			}
			time.Sleep(time.Second)
		}
	}
	return p
}