package client

import (
	"math/rand"
	dt "GO-DAG/DataTypes"
	"time"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/consensus"
	"GO-DAG/storage"
	"crypto/ecdsa"
	//"fmt"
)


func GenerateMessage(b []byte, signature []byte) []byte {
	var l uint32
	l = uint32(len(b))
	var magic_number uint32
	magic_number = 1
	payload := append(b,signature...)
	payload = append(serialize.EncodeToBytes(l),payload...)
	payload = append(serialize.EncodeToBytes(magic_number),payload...)
	return payload
}


func BroadcastTransaction(b []byte, p dt.Peers) {
	p.Mux.Lock() // lock for thr Peers datastructure
	for _,conn := range p.Fds {
		conn.Write(b)
	}
	p.Mux.Unlock()
}

func GenerateSignature(b []byte, PrivateKey *ecdsa.PrivateKey) []byte {
	hash := Crypto.Hash(b)
	signature := Crypto.Sign(hash[:],PrivateKey)
	return signature
}

func SimulateClient(p dt.Peers, PrivateKey *ecdsa.PrivateKey, dag *dt.DAG) {

	var tx dt.Transaction

	for {
		tx.Timestamp = time.Now().Unix()
		tx.Value = rand.Float64()
		copy(tx.From[:],Crypto.SerializePublicKey(&PrivateKey.PublicKey))
		dag.Mux.Lock()
		copy(tx.LeftTip[:],Crypto.DecodeToBytes(consensus.GetTip(dag,0.01)))
		copy(tx.RightTip[:],Crypto.DecodeToBytes(consensus.GetTip(dag,0.01)))
		dag.Mux.Unlock()
		Crypto.PoW(&tx,2)
		buffer := serialize.SerializeData(tx)
		sign := GenerateSignature(buffer,PrivateKey)
		msg := GenerateMessage(buffer,sign)
		storage.AddTransaction(dag,tx,sign)
		BroadcastTransaction(msg,p)
		time.Sleep(time.Second)
	}
}