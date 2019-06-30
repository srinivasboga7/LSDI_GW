package discovery

import (
	"io/ioutil"
	"strings"
	"net"
	dt "GO-DAG/DataTypes/datatypes.go"
)

func GetIps(filename string) []string {
	//Returns slice which has IPs as strings
	//Currently getting from file
	//Later get from DNS server or anything else
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	dataslice := strings.Split(string(bytes),"\n")
	dataslice = dataslice[:len(dataslice)-1]         //Delete last entry from slice as it will be empty
	return dataslice
}

func ConnectToServer(ips []string)(dt.Peers) {
	//Takes a map of ips to socket file descriptors of tcp connections
	var p dt.Peers
	for _,ip := range ips {
		tcpAddr, _ := net.ResolveTCPAddr("tcp4", ip)
		conn, _ := net.DialTCP("tcp", nil, tcpAddr)    //BLocking call
		p.Fds[ip] = conn
	}
	return p
}