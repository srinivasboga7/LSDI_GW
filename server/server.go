package server

import (
	// "fmt"
	// "time"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"encoding/binary"
	"net"
	//log "GO-DAG/logdump"
	"GO-DAG/Pow"
	"GO-DAG/storage"
	"encoding/json"
	"fmt"
	"strings"
)

type Server struct {
	Peers *storage.Peers
	Dag   *storage.DAG
}

func GetKeys(Graph map[string]storage.Vertex) []string {
	var keys []string
	for k := range Graph {
		keys = append(keys, k)
	}
	return keys
}

func (srv *Server) HandleConnection(connection net.Conn) {
	// each connection is handled in a seperate go routine
	addr := connection.RemoteAddr().String()
	ip := addr[:strings.IndexByte(addr, ':')]
	srv.Peers.Mux.Lock()
	if _, ok := srv.Peers.Fds[ip]; !ok {
		c, e := net.Dial("tcp", ip+":9000")
		if e != nil {
			//log.Println("CONNECTION UNSUCCESSFUL")
			fmt.Println()
		} else {
			srv.Peers.Fds[ip] = c
			//log.Println("CONNECTION REQUEST FROM A NEW PEER")
			fmt.Println()
		}
	}
	srv.Peers.Mux.Unlock()
	for {
		var buf []byte
		buf1 := make([]byte, 8) // reading the header
		headerLen, err := connection.Read(buf1)
		magicNumber := binary.LittleEndian.Uint32(buf1[:4])
		// specifies the type of the message
		if magicNumber == 1 {
			if headerLen < 8 {
				//log.Println("message broken")
				fmt.Println()
			} else {
				length := binary.LittleEndian.Uint32(buf1[4:8])
				buf2 := make([]byte, length+72)
				l, _ := connection.Read(buf2)
				buf = append(buf1, buf2[:l]...)
			}
		} else if magicNumber == 2 {
			buf = buf1
		} else if magicNumber == 3 {
			buf2 := make([]byte, 36)
			l, _ := connection.Read(buf2)
			buf = append(buf1, buf2[:l]...)
		}
		if err != nil {
			// Remove from the list of the peer
			srv.Peers.Mux.Lock()
			delete(srv.Peers.Fds, ip)
			srv.Peers.Mux.Unlock()
			// fmt.Println(err)
			break
		}
		if len(buf) > 0 {
			srv.HandleRequests(connection, buf, ip)
		}
	}
	defer connection.Close()
}

func (srv *Server) HandleRequests(connection net.Conn, data []byte, IP string) {
	magicNumber := binary.LittleEndian.Uint32(data[:4])
	if magicNumber == 1 {
		tx, sign := serialize.DeserializeTransaction(data[4:])
		if ValidTransaction(tx, sign) {
			// maybe wasting verifying duplicate transactions,
			// instead verify signatures and PoW while tip selection
			ok := storage.AddTransaction(srv.Dag, tx, sign, data[4:])
			if ok > 0 {
				//log.DefaultPrint("===========================================================================")
				//fmt.Println()
				//log.Println("RECIEVED TRANSACTION ")
				//log.DefaultPrintBlue(Crypto.EncodeToHex(tx.TxID[:]))
				//fmt.Println()
				//time.Sleep(time.Second)
				//log.Println("TRANSACTION VERIFIED ")
				//log.DefaultPrintBlue(Crypto.EncodeToHex(tx.TxID[:]))
				//fmt.Println()
				//time.Sleep(time.Second)
				//log.Println("TRANSACTION ADDED TO DAG ")
				//log.DefaultPrintBlue(Crypto.EncodeToHex(tx.TxID[:]))
				//fmt.Println()
				//time.Sleep(time.Second)
				//log.Println("FORWARDING TRANSACTION TO OTHER PEERS ")
				//log.DefaultPrintBlue(Crypto.EncodeToHex(tx.TxID[:]))
				//fmt.Println()
				//log.DefaultPrint("===========================================================================")
				//fmt.Println()
				srv.ForwardTransaction(data, IP)
			}
		}
	} else if magicNumber == 2 {
		//log.Println("PEER REQUESTING TRANSACTION TO SYNC")
		fmt.Println()
		// request to give the hashes of tips
		srv.Dag.Mux.Lock()
		ser, _ := json.Marshal(GetKeys(srv.Dag.Graph))
		srv.Dag.Mux.Unlock()
		connection.Write(ser)
	} else if magicNumber == 3 {
		// request to give transactions based on tips
		hash := data[4:36]
		str := Crypto.EncodeToHex(hash)
		srv.Dag.Mux.Lock()
		if val, ok := srv.Dag.Graph[str]; ok {
			tx, sign := val.Tx, val.Signature
			srv.Dag.Mux.Unlock()
			reply := serialize.SerializeData(tx)
			var l uint32
			l = uint32(len(reply))
			reply = append(reply, sign...)
			reply = append(serialize.EncodeToBytes(l), reply...)
			reply = append(serialize.EncodeToBytes("1"), reply...)
			connection.Write(reply)
		} else {
			reply := serialize.EncodeToBytes("0")
			connection.Write(reply)
		}
	} else if magicNumber == 4 {
		// Not relevant but given hash it responds with hashes of neighbours
		hash := data[4:36]
		str := Crypto.EncodeToHex(hash)
		srv.Dag.Mux.Lock()
		reply, _ := json.Marshal(srv.Dag.Graph[str].Neighbours)
		srv.Dag.Mux.Unlock()
		connection.Write(reply)
	} else {
		//log.Println("FAILED REQUEST")
		fmt.Println()
	}
}

func ValidTransaction(t storage.Transaction, signature []byte) bool {
	// check the signature
	s := serialize.SerializeData(t)
	SerialKey := t.From
	PublicKey := Crypto.DeserializePublicKey(SerialKey[:])
	h := Crypto.Hash(s)
	sigVerify := Crypto.Verify(signature, PublicKey, h[:])
	if sigVerify == false {
		//log.Println("INVALID SIGNATURE")
		//fmt.Println()
	}
	return sigVerify && Pow.VerifyPoW(t, 4)
	//return Crypto.VerifyPoW(t,2)
}

func (srv *Server) ForwardTransaction(t []byte, IP string) {
	// sending the transaction to the peers excluding the one it came from
	////log.Println("Relayed to other peers")
	srv.Peers.Mux.Lock()
	for k, conn := range srv.Peers.Fds {
		if k != IP {
			conn.Write(t)
		}
	}
	srv.Peers.Mux.Unlock()
}

func (srv *Server) StartServer() {
	listener, _ := net.Listen("tcp", ":9000")
	for {
		conn, _ := listener.Accept()
		go srv.HandleConnection(conn)
		// go routine executes concurrently
	}
}
