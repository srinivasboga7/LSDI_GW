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

func Getkeys(m map[p2p.PeerID] struct{}) []p2p.PeerID {
	keys := make([]p2p.PeerID, 0, len(m))
    for k := range m {
        keys = append(keys, k)
	}
	return keys
}


func main() {
	rand.Seed(time.Now().UnixNano())
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
		ch, shsch, shtxch = node.NewBootstrap(&ID, &dag)
	} else {
		ch, shsch, shtxch = node.New(&ID, &dag)
	}
	var cli client.Client
	cli.PrivateKey = PrivateKey
	cli.Send = ch
	cli.DAG = &dag
	go cli.SimulateClient()
	// ShardRequesters p2p.PeerID[]
	ShardRequesters map[p2p.PeerID] struct{}
	lastShardSignal = make(dt.ShardSignal)
	var localShardNo int32
	unverifiedShardlist = make(p2p.PeerID)
	shardingActive = uint32
	shardingActive = 0
	for {
		select {
		case signal := <-shsch:
			if lastShardSignal != signal {
				lastShardSignal = signal
				Shardingtx, err := sh.MakeShardingtx(Puk, signal, sign)
				localShardNo = Shardingtx.ShardNo
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
				for _,s := range unverifiedShardlist {
					if _,ok := ShardRequesters[Sender]; !ok {
						ShardRequesters[Sender] = Sender
					}
					if len(ShardRequesters) > 7 && *myShardNo != -1 {
						log.Println("Threshold broken")
						NewShardList := Getkeys(ShardRequesters)
						orderToConnect := rand.Perm(len(NewShardList))
						ShardRequesters map[p2p.PeerID] struct{}
						DisconnectOldConn()
						MakeNewConn(orderToConnect, NewShardList)
						localShardNo = -1
						unverifiedShardlist = make(p2p.PeerID)
					}
				}
			}
		case tx, Sender := <-shtxch:
			if tx.Identifier != lastShardSignal.Identifier {
				if localShardNo == -1 {
					unverifiedShardlist = append(unverifiedShardlist, Sender)
				} else if _,ok := ShardRequesters[Sender]; !ok && tx.ShardNo == localShardNo {
					ShardRequesters[Sender] = Sender
				}
				if len(ShardRequesters) > 7 && localShardNo != -1 {
					log.Println("Threshold broken")
					NewShardList := Getkeys(ShardRequesters)
					orderToConnect := rand.Perm(len(NewShardList))
					ShardRequesters map[p2p.PeerID] struct{}
					DisconnectOldConn()
					MakeNewConn(orderToConnect, NewShardList)
					localShardNo = -1
				}
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