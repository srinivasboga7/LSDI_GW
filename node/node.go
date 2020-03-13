package node

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	sh "GO-DAG/sharding"
	"GO-DAG/storage"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	f, _ = os.Create("logFile.txt")
)

// New ...
func New(hostID *p2p.PeerID, dag *dt.DAG, PrivKey Crypto.PrivateKey) chan p2p.Msg {

	var srv p2p.Server
	srv.HostID = *hostID
	srv.BroadcastMsg = make(chan p2p.Msg)
	srv.NewPeer = make(chan p2p.Peer)
	srv.RemovePeer = make(chan p2p.Peer)
	srv.ShardTransactions = make(chan dt.ShardTransaction)
	srv.ShardingSignal = make(chan dt.ShardSignal)
	srv.PrivateKey = PrivKey
	go srv.Run()
	go func() {
		for {
			p := <-srv.NewPeer
			go handle(&p, srv.BroadcastMsg, dag, srv.ShardingSignal, srv.ShardTransactions, srv.RemovePeer)
		}
	}()
	return srv.BroadcastMsg
}

func handleMsg(msg p2p.Msg, send chan p2p.Msg, dag *dt.DAG, p *p2p.Peer, ShardSignalch chan dt.ShardSignal, Shardtxch chan dt.ShardTransaction) {
	// check for transactions or request for transactions
	if msg.ID == 32 {
		// transaction
		tx, sign := serialize.Decode32(msg.Payload, msg.LenPayload)
		// log.Println(hex.EncodeToString(tx.LeftTip[:]))
		if validTransaction(msg.Payload, tx.From[:]) {
			tr := storage.AddTransaction(dag, tx, sign)
			if tr == 1 {
				send <- msg
				f.WriteString(fmt.Sprintf("%d\n", time.Now().Nanosecond()))
			} else if tr == 2 {
				var msg p2p.Msg
				msg.ID = 34
				if !storage.CheckifPresentDb(tx.LeftTip[:]) {
					msg.Payload = tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
				if !storage.CheckifPresentDb(tx.RightTip[:]) {
					msg.Payload = tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
			}
		}
	} else if msg.ID == 34 {
		// request for transaction
		hash := msg.Payload
		v, ok := dag.Graph[Crypto.EncodeToHex(hash)]
		if ok {
			tx := v.Tx
			sign := v.Signature
			var respMsg p2p.Msg
			respMsg.ID = 33
			respMsg.Payload = append(serialize.Encode32(tx), sign...)
			respMsg.LenPayload = uint32(len(respMsg.Payload))
			p.Send(respMsg)
		}
	} else if msg.ID == 33 {
		tx, sign := serialize.Decode32(msg.Payload, msg.LenPayload)
		if validTransaction(msg.Payload, tx.From[:]) {
			tr := storage.AddTransaction(dag, tx, sign)
			if tr == 2 {
				var msg p2p.Msg
				msg.ID = 34
				if !storage.CheckifPresentDb(tx.LeftTip[:]) {
					msg.Payload = tx.LeftTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
				if !storage.CheckifPresentDb(tx.RightTip[:]) {
					msg.Payload = tx.RightTip[:]
					msg.LenPayload = uint32(len(msg.Payload))
					p.Send(msg)
				}
			}
		}
	} else if msg.ID == 35 { //Shard signal
		signal, _ := serialize.Decode35(msg.Payload, msg.LenPayload)
		select {
		case ShardSignalch <- signal:
			send <- msg
		default:
		}
	} else if msg.ID == 36 { //Sharding tx from other nodes
		tx, sign := serialize.Decode36(msg.Payload, msg.LenPayload)
		if sh.VerifyShardTransaction(tx, sign, 4) {
			Shardtxch <- tx
		}
	}
}

// read the messages and handle
func handle(p *p2p.Peer, send chan p2p.Msg, dag *dt.DAG, ShardSignalch chan dt.ShardSignal, Shardtxch chan dt.ShardTransaction, errChan chan p2p.Peer) {
	for {
		msg, err := p.GetMsg()
		if err != nil {
			errChan <- *p
			log.Println(err)
			break
		}
		go handleMsg(msg, send, dag, p, ShardSignalch, Shardtxch)
	}
}

func validTransaction(msg []byte, PublicKey []byte) bool {

	sTx := msg[:len(msg)-72]
	sign := msg[len(msg)-72:]
	h := Crypto.Hash(sTx)
	return Crypto.Verify(sign, Crypto.DeserializePublicKey(PublicKey), h[:])
}

func getAllKeys(dag *dt.DAG) []string {

	var hashes []string
	dag.Mux.Lock()
	for k := range dag.Graph {
		hashes = append(hashes, k)
	}
	dag.Mux.Unlock()
	return hashes
}
