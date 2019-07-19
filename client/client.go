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


func GenerateMessage(b []byte, PrivateKey *ecdsa.PrivateKey) []byte {
	hash := Crypto.Hash(b)
	signature := Crypto.Sign(hash[:],PrivateKey)
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

func SimulateClient(p dt.Peers, PrivateKey *ecdsa.PrivateKey, dag dt.DAG) {

	var tx dt.Transaction

	for {
		tx.Timestamp = time.Now().Unix()
		tx.Value = rand.Float64()
		copy(tx.From[:],Crypto.SerializePublicKey(&PrivateKey.PublicKey))
		copyLedger := dag
		copy(tx.LeftTip[:],Crypto.DecodeToBytes(consensus.GetTip(copyLedger,0.1)))
		copy(tx.RightTip[:],Crypto.DecodeToBytes(consensus.GetTip(copyLedger,0.1)))
		Crypto.PoW(&tx,4)
		//fmt.Println("Selected Tips")
		storage.AddTransaction(dag,tx)
		buffer := serialize.SerializeData(tx)
		msg := GenerateMessage(buffer,PrivateKey)
		BroadcastTransaction(msg,p)
		time.Sleep(time.Second)
	}
}