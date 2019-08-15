package server

import(
	"fmt"
	"net"
	"time"
	"encoding/binary"
	"encoding/json"
	dt "GO-DAG/DataTypes"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"strings"
)

func HandleConnection(connection net.Conn, p dt.Peers, dag *dt.DAG) {
	// each connection is handled in a seperate go routine
	timeoutDuration := 2700*time.Second
	for {
		connection.SetReadDeadline(time.Now().Add(timeoutDuration))
		buf := make([]byte,1024)
		l, err := connection.Read(buf)
		if err != nil {
			// Remove from the list of the peer
			fmt.Println(err)
			break
		}
		addr := connection.RemoteAddr().String()
		ip := addr[:strings.IndexByte(addr,':')] // seperate the port to get only IP
		go HandleRequests(connection,buf[:l],ip,p,dag)
	}
	defer connection.Close()
}


func HandleRequests (connection net.Conn,data []byte, IP string, p dt.Peers, dag *dt.DAG) {
	magic_number := binary.LittleEndian.Uint32(data[:4])
	if magic_number == 1 {
		tx,sign := serialize.DeserializeTransaction(data[4:])
		if ValidTransaction(tx,sign) { // maybe wasting verifying duplicate transactions, 
			// instead verify signatures and PoW while tip selection
			if storage.AddTransaction(dag,tx,sign) {
				//ForwardTransaction(data,IP,p) // Duplicates are not forwarded
			}
		}
	} else if magic_number == 2 {
		var req dt.RequestTx
		json.Unmarshal(data[4:],&req)
		dag.Mux.RLock()
		tx := dag.Graph[req.Hash]
		dag.Mux.RUnlock()
		reply,_ := json.Marshal(tx)
		connection.Write(reply)
	} else if magic_number == 3 {
		var req dt.RequestConnection
		json.Unmarshal(data[4:],&req)
		IP := req.IP
		peer,_ := net.Dial("tcp",IP+":9000")
		p.Mux.Lock()
		p.Fds[IP] = peer
		p.Mux.Unlock()
	} else {
		fmt.Println("Invalid message from other peer")
	}
}



func ValidTransaction(t dt.Transaction, signature []byte) bool {
	// check the signature
	s := serialize.SerializeData(t)
	SerialKey := t.From
	PublicKey := Crypto.DeserializePublicKey(SerialKey[:])
	h := Crypto.Hash(s)
	sigVerify := Crypto.Verify(signature,PublicKey,h[:])
	if sigVerify == false {
		fmt.Println("Invalid signature")
	}
	return sigVerify && Crypto.VerifyPoW(t,2)
	//return Crypto.VerifyPoW(t,2)
}


func ForwardTransaction(t []byte, IP string, p dt.Peers) {
	// sending the transaction to the peers excluding the one it came from
	//fmt.Println("Relayed to other peers")	
	p.Mux.Lock()
	for k,conn := range p.Fds {
		if k != IP {
			conn.Write(t)
		} 
	}
	p.Mux.Unlock()
}


func StartServer(p dt.Peers, dag *dt.DAG) {
	listener, _ := net.Listen("tcp",":9000")
	for {
		conn, _ := listener.Accept()
 		go HandleConnection(conn,p,dag) // go routine executes concurrently

	}
	defer listener.Close()
}