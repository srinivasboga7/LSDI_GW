package sync

import(
	dt "GO-DAG/datatypes"
	"GO-DAG/storage"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"net"
	"encoding/json"
)

//copyDAG Sync dag copy - Copies latest transactions from peer and constructs dag
func copyDAG(dag *dt.DAG, p *dt.Peers, conn net.Conn) {
	var magicNumber uint32
	magicNumber = 2
	b := serialize.EncodeToBytes(magicNumber)
	var txs []string
	p.Mux.Lock()
	conn.Write(b)
	var ser []byte
	for { 
		buf := make([]byte,1024)
		l,_ := conn.Read(buf)
		ser = append(ser,buf[:l]...)
		if l < 1024 {
			break
		}
	}
	json.Unmarshal(ser,&txs)
	p.Mux.Unlock()
	magicNumber = 3
	num := serialize.EncodeToBytes(magicNumber)
	var v string
	for _,v = range txs {
		hash := Crypto.DecodeToBytes(v)
		hash = append(num,hash...)
		conn.Write(hash)
		bufCheck := make([]byte,1)
		l,_ := Peers.Fds[r].Read(bufCheck)
		if Crypto.EncodeToHex(bufCheck) == "1" {
	    	buf := make([]byte,1024)
			length,_ := Peers.Fds[r].Read(buf)
			Peers.Mux.Unlock()
			tx,sign := serialize.DeserializeTransaction(buf[:length])
			var vertex dt.Vertex
			vertex.Tx = tx
			vertex.Signature = sign
			var tip [32]byte 
			dag.Mux.Lock()
			if tx.LeftTip == tip && tx.RightTip == tip {
				dag.Genisis = v
				// log.Println("Genisis Transaction")
			}
			dag.Graph[v] = vertex
			dag.Mux.Unlock()
		}
		dag.Mux.Unlock()
	}
	constructDAG(dag)
}

//constructDAG Constructs DAG structure from transctions using tips, used in the copyDAG function
func constructDAG(dag *dt.DAG) {
	for h,vertex := range dag.Graph {
		tx := vertex.Tx
		left := Crypto.EncodeToHex(tx.LeftTip[:])
		right := Crypto.EncodeToHex(tx.RightTip[:])
		if left != right {
			if _,ok := dag.Graph[left] ; ok {
				l := dag.Graph[left]
				l.Neighbours = append(l.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			}
			if _,ok := dag.Graph[right] ; ok {
				r := dag.Graph[right]
				r.Neighbours = append(r.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])] = r
			}
		} else {
			if _,ok := dag.Graph[left] ; ok {
				l := dag.Graph[left]
				l.Neighbours = append(l.Neighbours,h)
				dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			}
		}
	}

	// sync with the orphaned Transactions
	for _,vertices := range storage.OrphanedTransactions {
		for _, v := range vertices {
			storage.AddTransaction(dag,v.Tx,v.Signature)
		}
	}
}