package server

import(
	"fmt"
	"net"
	dt "GO-DAG/DataTypes"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"strings"
)

func HandleConnection(connection net.Conn, p dt.Peers, c chan dt.Transaction) {
	// each connection is handled in a seperate go routine
	for {
		buf := make([]byte,1024)
		len, err := connection.Read(buf)
		if err != nil {
			// Remove from the list of the peer
			fmt.Println(err)
			break
		}
		addr := connection.RemoteAddr().String()
		ip := addr[:strings.IndexByte(addr,':')] // seperate the port to get only IP
		//fmt.Println(ip)
		HandleRecievedData(buf[:len],ip,p,c)
	}
	defer connection.Close()
}


func HandleRecievedData (data []byte, IP string, p dt.Peers,c chan dt.Transaction) {
	tx,sign := serialize.Deserialize(data)
	if ValidTransaction(tx,sign) {
		fmt.Println("Valid Transaction")
		ForwardTransaction(data,IP,p)
		AddTransaction(tx,c)
	}
}


func ValidTransaction(t dt.Transaction, signature []byte) bool {
	// check the signature
	SerialKey := t.From
	PublicKey := Crypto.DeserializePublicKey(SerialKey[:])
	s := serialize.SerializeData(t)
	h := Crypto.Hash(s)
	return Crypto.Verify(signature,PublicKey,h[:])
}



func AddTransaction(t dt.Transaction, c chan dt.Transaction) {
	// Add the transaction to the DAG
	c <- t
	return
}

func ForwardTransaction(t []byte, IP string, p dt.Peers) {
	// sending the transaction to the peers excluding the one it came from	
	p.Mux.Lock()
	for k,conn := range p.Fds {
		if k != IP {
			conn.Write(t)
		} 
	}
	p.Mux.Unlock()
}


func StartServer(p dt.Peers, c chan dt.Transaction) {
	listener, _ := net.Listen("tcp",":9000")
	for {
		conn, _ := listener.Accept()
 		go HandleConnection(conn,p,c) // go routine executes concurrently

	}
	defer listener.Close()
}