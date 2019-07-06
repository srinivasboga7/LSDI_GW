package client

import (
	"math/rand"
	dt "GO-DAG/DataTypes"
	"time"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"crypto/ecdsa"
)


func GenerateMessage(b []byte, PrivateKey *ecdsa.PrivateKey) []byte {
	hash := Crypto.Hash(b)
	signature := Crypto.Sign(hash[:],PrivateKey)
	var l uint32
	l = uint32(len(b))
	payload := append(b,signature...)
	payload = append(serialize.EncodeToBytes(l),payload...)
	return payload
}


func BroadcastTransaction(b []byte, p dt.Peers) {
	p.Mux.Lock() // lock for thr Peers datastructure
	for _,conn := range p.Fds {
		conn.Write(b)
	}
	p.Mux.Unlock()
}


func SimulateClient(p dt.Peers, PrivateKey *ecdsa.PrivateKey) {

	var tx dt.Transaction

	for {
		tx.Timestamp = time.Now().Unix()
		tx.Value = rand.Float64()
		copy(tx.From[:],Crypto.SerializePublicKey(&PrivateKey.PublicKey))
		buffer := serialize.SerializeData(tx)
		msg := GenerateMessage(buffer,PrivateKey)
		BroadcastTransaction(msg,p)
		time.Sleep(2*time.Second)
	}
}