package client

import (
	"math/rand"
	"io/ioutil"
	"strings"
	"net"
	dt "GO-DAG/DataTypes"
	"time"
	"encoding/binary"
	"bytes"
)

func SerializeData(t dt.Transaction) []byte {
	// iterating over a struct is painful in golang
	var b []byte
	v := reflect.ValueOf(&t).Elem()
	for i := 0; i < v.NumField() ;i++ {
		value := v.Field(i)
		b = append(b,EncodeToBytes(value.Interface())...)
	}
	return b
}

func EncodeToBytes(x interface{}) []byte {
	switch t := x.(type) {
	case string :
		return []byte(x)
	default :
		buf := new(bytes.Buffer)
		err := binary.Write(buf,binary.LittleEndian, v)
		return buf.Bytes()
	}
}


func BroadcastTransaction(b []byte, p dt.Peers) {
	p.Mux.Lock()
	buf := b
	for _,conn := range p.Fds {
		conn.Write(buf)
	}
	p.Mux.Unlock()
}

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
	//Takes a slice of ips and returns sock file descriptors of tcp connections
	var p dt.Peers
	for _,ip := range ips {
		tcpAddr, _ := net.ResolveTCPAddr("tcp4", ip)
		conn, _ := net.DialTCP("tcp", nil, tcpAddr)    //BLocking call
		p.Fds = append(p.Fds,conn)
	}
	return p
}

func SimulateClient(p dt.Peers) {

	var tx dt.Transaction

	for {
		tx.Timestamp = time.Now().Unix()
		tx.Value = rand.Float64()
		buffer := SerializeData(tx)
		BroadcastTransaction(buffer,p)
		time.Sleep(2*time.Second)
	}
}