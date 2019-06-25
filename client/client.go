package client

import (
	//"crypto/ecdsa"
	"math/rand"
	dt "GO-DAG/DataTypes"
	//"net"
	"time"
	"encoding/gob"
	"bytes"
)

func SerializeData(t dt.Transaction) []byte {
	// too much overhead has to change
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	e.Encode(t)
	buf := b.Bytes()
	return buf
}

func BroadcastTransaction(b []byte, p dt.Peers) {
	p.Mux.Lock()
	buf := b
	for _,conn := range p.Fds {
		conn.Write(buf)
	}
	p.Mux.Unlock()

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