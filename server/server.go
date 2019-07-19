package server

import(
	"fmt"
	"net"
	"encoding/binary"
	dt "GO-DAG/DataTypes"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"strings"
)

func HandleConnection(connection net.Conn, p dt.Peers, dag dt.DAG) {
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
		HandleRequests(buf[:len],ip,p,dag)
	}
	defer connection.Close()
}


func HandleRequests (data []byte, IP string, p dt.Peers, dag dt.DAG) {
	magic_number := binary.LittleEndian.Uint32(data[:4])
	if magic_number == 1 {
		tx,sign := serialize.DeserializeTransaction(data[4:])
		if ValidTransaction(tx,sign) {
			//fmt.Println("Valid Transaction")
			if storage.AddTransaction(dag,tx) {
				ForwardTransaction(data,IP,p) // Duplicates are not forwarded
			}
		}
	}
}



func ValidTransaction(t dt.Transaction, signature []byte) bool {
	// check the signature
	SerialKey := t.From
	PublicKey := Crypto.DeserializePublicKey(SerialKey[:])
	s := serialize.SerializeData(t)
	h := Crypto.Hash(s)
	return Crypto.Verify(signature,PublicKey,h[:]) && Crypto.VerifyPoW(t,4)
}


func ForwardTransaction(t []byte, IP string, p dt.Peers) {
	// sending the transaction to the peers excluding the one it came from
	fmt.Println("Relayed to other peers")	
	p.Mux.Lock()
	for k,conn := range p.Fds {
		if k != IP {
			conn.Write(t)
		} 
	}
	p.Mux.Unlock()
}


func StartServer(p dt.Peers, dag dt.DAG) {
	listener, _ := net.Listen("tcp",":9000")
	for {
		conn, _ := listener.Accept()
 		go HandleConnection(conn,p,dag) // go routine executes concurrently

	}
	defer listener.Close()
}