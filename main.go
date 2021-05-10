package main

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"GO-DAG/client"
	"GO-DAG/gateway"
	"GO-DAG/node"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	var PrivateKey Crypto.PrivateKey
	// checking for a private key if provided
	if Crypto.CheckForKeys() {
		PrivateKey = Crypto.LoadKeys()
	} else {
		PrivateKey = Crypto.GenerateKeys()
	}
	var ID p2p.PeerID
	ID.PublicKey = Crypto.SerializePublicKey(&PrivateKey.PublicKey)
	var dag dt.DAG
	v := constructGenisis()
	genisisHash := Crypto.Hash(serialize.Encode32(v.Tx))
	dag.Graph = make(map[string]dt.Vertex)
	dag.Graph[Crypto.EncodeToHex(genisisHash[:])] = v
	var ch chan p2p.Msg
	storageCh := make(chan dt.ForwardTx, 20)
	dag.StorageCh = storageCh
	ch = node.New(&ID, &dag, PrivateKey)
	// initializing the storage layer
	var st storage.Server
	st.DAG = &dag
	st.ForwardingCh = ch
	st.ServerCh = storageCh
	go st.Run()
	var cli client.Client
	cli.PrivateKey = PrivateKey
	cli.Send = ch
	cli.DAG = &dag
	// API for accpeting data from the sensors
	go gateway.RunAPI()
	// API for monitoring metrics and also generating transactions from the hash value
	cli.RunAPI()
}

func constructGenisis() dt.Vertex {
	var tx dt.Transaction
	tx.Hash = Crypto.Hash([]byte("IOT BLOCKCHAIN GENISIS"))
	var v dt.Vertex
	v.Tx = tx
	v.Signature = make([]byte, 72)
	return v
}
