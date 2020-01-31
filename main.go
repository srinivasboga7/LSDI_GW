package main

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"GO-DAG/client"
	"GO-DAG/node"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"os"
)

func main() {
	var PrivateKey Crypto.PrivateKey
	if Crypto.CheckForKeys() {
		PrivateKey = Crypto.LoadKeys()
	} else {
		PrivateKey = Crypto.GenerateKeys()
	}
	var ID p2p.PeerID
	ID.PublicKey = Crypto.SerializePublicKey(&PrivateKey.PublicKey)
	var dag dt.DAG
	v := constructGenisis()
	genisisHash := Crypto.Hash(serialize.Encode(v.Tx))
	dag.Graph = make(map[string]dt.Vertex)
	dag.Graph[Crypto.EncodeToHex(genisisHash[:])] = v
	var ch chan p2p.Msg
	if os.Args[1] == "b" {
		ch = node.NewBootstrap(ID, &dag)
	} else {
		ch = node.New(ID, &dag)
	}
	var cli client.Client
	cli.PrivateKey = PrivateKey
	cli.Send = ch
	cli.DAG = &dag
	cli.SimulateClient()
}

func constructGenisis() dt.Vertex {
	var tx dt.Transaction
	tx.Hash = Crypto.Hash([]byte("IOT BLOCKCHAIN GENISIS"))
	var v dt.Vertex
	v.Tx = tx
	v.Signature = make([]byte, 72)
	return v
}
