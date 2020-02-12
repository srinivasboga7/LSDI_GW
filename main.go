package main

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"GO-DAG/client"
	"GO-DAG/node"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"os"
	"time"
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
		ch, shsch, shtxch = node.NewBootstrap(ID, &dag)
	} else {
		ch, shsch, shtxch = node.New(ID, &dag)
	}
	var cli client.Client
	cli.PrivateKey = PrivateKey
	cli.Send = ch
	cli.DAG = &dag
	go cli.SimulateClient()
	for {
		select {
		case signal := <-shsch:
			Shardingtx, err := sh.MakeShardingtx(Puk, signal, sign)
			*localShardNo = Shardingtx.ShardNo
			if err != nil {
				fmt.Println(err)
				return
			}
			sendStream := serialize.Encode(Shardingtx)
			var msg p2p.Msg
			msg.ID = 36
			msg.Payload = sendStream
			msg.LenPayload = uint32(len(msg.Payload))
			send <- msg
			log.Println("Sharding broadcast..")
			log.Println(Shardingtx)
		case Shardingtx := <-shtxch:
			if len(NewShardList) > 7 && *myShardNo != -1 {
				log.Println("Threshold broken")
				*myShardNo = -1
				OldShardList = NewShardList
				NewShardList = make([]string)
				rand.Seed(time.Now().UnixNano())
				orderToConnect := rand.Perm(len(OldShardList))
				DisconnectOldConn()
				MakeNewConn(orderToConnect)
			}
		}
	}
}

func constructGenisis() dt.Vertex {
	var tx dt.Transaction
	tx.Hash = Crypto.Hash([]byte("IOT BLOCKCHAIN GENISIS"))
	tx.Timestamp = time.Now().UnixNano()
	var v dt.Vertex
	v.Tx = tx
	return v
}
