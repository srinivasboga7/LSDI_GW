package server

import(
	"fmt"
	"net"
	//"time"
	"encoding/binary"
	dt "GO-DAG/DataTypes"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"encoding/json"
	"strings"
)

type Server struct {
	Peers *dt.Peers
	Dag *dt.DAG
}

func GetKeys(Graph map[string]dt.Node) []string {
	var keys []string
	for k,v := range Graph {
		if len(v.Neighbours) < 2 {
			keys = append(keys,k)
		}
	}
	return keys
}


func (srv *Server) HandleConnection(connection net.Conn) {
	// each connection is handled in a seperate go routine
	addr := connection.RemoteAddr().String()
	ip := addr[:strings.IndexByte(addr,':')]
	srv.Peers.Mux.Lock()
	if _,ok := srv.Peers.Fds[ip] ; !ok {
		c,e := net.Dial("tcp",ip+":9000")
		if e != nil {
			fmt.Println("Connection Unsuccessful")
		} else {
			srv.Peers.Fds[ip] = c
			fmt.Println("Connection Successful")
		}
	}
	srv.Peers.Mux.Unlock()
	for {
		var buf []byte
		buf1 := make([]byte,8) // reading the header 
		_,err := connection.Read(buf1)
		magic_number := binary.LittleEndian.Uint32(buf1[:4]) 
		// specifies the type of the message
		if magic_number == 1 { 
			length := binary.LittleEndian.Uint32(buf1[4:8])
			buf2 := make([]byte,length+72)
			l,_ := connection.Read(buf2)
			buf = append(buf1,buf2[:l]...)
		} else if magic_number == 2 {
			buf = buf1
		} else if magic_number == 3 {
			buf2 := make([]byte,36)
			l,_ := connection.Read(buf2)
			buf = append(buf1,buf2[:l]...)
		}
		if err != nil {
			// Remove from the list of the peer
			srv.Peers.Mux.Lock()
			delete(srv.Peers.Fds,ip)
			srv.Peers.Mux.Unlock()
			fmt.Println(err)
			break
		}
		if len(buf) > 0 {
			go srv.HandleRequests(connection,buf,ip)
		}
	}
	defer connection.Close()
}


func (srv *Server)HandleRequests (connection net.Conn,data []byte, IP string) {
	magic_number := binary.LittleEndian.Uint32(data[:4])
	if magic_number == 1 {
		tx,sign := serialize.DeserializeTransaction(data[4:])
		if ValidTransaction(tx,sign) { 
			// maybe wasting verifying duplicate transactions, 
			// instead verify signatures and PoW while tip selection
			if storage.AddTransaction(srv.Dag,tx,sign) {
				srv.ForwardTransaction(data,IP)
			}
		}
	} else if magic_number == 2 {
		// request to give the hashes of tips 
		srv.Dag.Mux.Lock()
		ser,_ := json.Marshal(GetKeys(srv.Dag.Graph))
		srv.Dag.Mux.Unlock()
		connection.Write(ser)
	} else if magic_number == 3 {
		// request to give transactions based on tips
		hash := data[4:36]
		str := Crypto.EncodeToHex(hash)
		srv.Dag.Mux.Lock()
		tx,sign := srv.Dag.Graph[str].Tx,srv.Dag.Graph[str].Signature
		srv.Dag.Mux.Unlock()
		reply := serialize.SerializeData(tx)
		var l uint32
		l = uint32(len(reply))
		reply = append(reply,sign...)
		reply = append(serialize.EncodeToBytes(l),reply...)
		connection.Write(reply)
	} else if magic_number == 4{
		// Not relevant but given hash it responds with hashes of neighbours
		hash := data[4:36]
		str := Crypto.EncodeToHex(hash)
		srv.Dag.Mux.Lock()
		reply,_ := json.Marshal(srv.Dag.Graph[str].Neighbours)
		srv.Dag.Mux.Unlock()
		connection.Write(reply) 
	} else {
		fmt.Println("Failed Request")
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


func (srv *Server)ForwardTransaction(t []byte, IP string) {
	// sending the transaction to the peers excluding the one it came from
	//fmt.Println("Relayed to other peers")	
	srv.Peers.Mux.Lock()
	for k,conn := range srv.Peers.Fds {
		if k != IP {
			conn.Write(t)
		} 
	}
	srv.Peers.Mux.Unlock()
}


func (srv *Server)StartServer() {
	listener, _ := net.Listen("tcp",":9000")
	for {
		conn, _ := listener.Accept()
 		go srv.HandleConnection(conn) 
 		// go routine executes concurrently
	}
	defer listener.Close()
}