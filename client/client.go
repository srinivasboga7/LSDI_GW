package client

import (
	"fmt"
	"encoding/binary"
	"encoding/json"
	dt "GO-DAG/DataTypes"
	"time"
	"net/http"
	"github.com/google/uuid"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/consensus"
	"GO-DAG/storage"
	log "GO-DAG/logdump" 
	"crypto/ecdsa"
	"bytes"
	"net"
)

// Client is declared to be used in main
type Client struct {
	PrivateKey *ecdsa.PrivateKey
	Dag *dt.DAG
	Peers *dt.Peers
}

type sensordata struct {
	Data string
	SensorName string
	//SensorID string
	SmID string
	Start int64
	//End int64
}

type postRequest struct {
	SerialData string
	TxID string
}

// GenerateMessage creates a byte slice to be sent over a network
// for a particular transaction.
func GenerateMessage(b []byte, signature []byte) []byte {
	var l uint32
	l = uint32(len(b))
	var magicNumber uint32
	magicNumber = 1
	payload := append(b,signature...)
	payload = append(serialize.EncodeToBytes(l),payload...)
	payload = append(serialize.EncodeToBytes(magicNumber),payload...)
	return payload
}

// BroadcastTransaction sends the transaction to all the peers
func BroadcastTransaction(b []byte, p *dt.Peers) {
	p.Mux.Lock()
	for _,conn := range p.Fds {
		conn.Write(b)
	}
	p.Mux.Unlock()
}

// GenerateSignature uses Crypto pkg to genearte a signature.
// It is a wrapper aroung Crypto.Sign function
func GenerateSignature(b []byte, PrivateKey *ecdsa.PrivateKey) []byte {
	hash := Crypto.Hash(b)
	signature := Crypto.Sign(hash[:],PrivateKey)
	return signature
}

// Copy generates the Deep copy of the DAG
func Copy(dag *dt.DAG) *dt.DAG {
	bytes,_ := json.Marshal(dag.Graph)
	var copyGraph map[string] dt.Vertex
	json.Unmarshal(bytes,&copyGraph)
	var copyDag dt.DAG
	copyDag.Graph = copyGraph
	copyDag.Genisis = dag.Genisis
	return &copyDag
}

func (cli *Client)createTransaction(data []byte, url string) {
	privateKey := cli.PrivateKey
	var tx dt.Transaction
	tx.Timestamp = time.Now().Unix()
	tx.Hash = Crypto.Hash(data)
	tx.TxID = uuid.New()
	var request postRequest 
	request.SerialData = string(data)
	request.TxID = Crypto.EncodeToHex(tx.TxID[:])
	serial,err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
	}
	// log.Println(request)
	b := bytes.NewReader(serial)
	log.Println("FORWARDING SMARTMETER DATA TO CLOUD")
	http.Post(url,"application/json",b)
	log.Println("CREATING TRANSACTION USING HASH OF SMARTMETER DATA")
	copy(tx.From[:],Crypto.SerializePublicKey(&privateKey.PublicKey))
	log.DefaultPrint("===========================================")
	log.Println("STARTING TIP SELECTION")
	cli.Dag.Mux.Lock()
	copyDag := cli.Dag
	copy(tx.LeftTip[:],Crypto.DecodeToBytes(consensus.GetTip(copyDag,0.01)))
	copy(tx.RightTip[:],Crypto.DecodeToBytes(consensus.GetTip(copyDag,0.01)))
	cli.Dag.Mux.Unlock()
	log.Println("TIP SELECTION COMPLETED")
	log.DefaultPrint("===========================================")
	log.Println("STARTING PoW")
	Crypto.PoW(&tx,4)
	log.Println("PoW FINISHED")
	log.DefaultPrint("===========================================")
	buffer := serialize.SerializeData(tx)
	sign := GenerateSignature(buffer,privateKey)
	msg := GenerateMessage(buffer,sign)
	storage.AddTransaction(cli.Dag,tx,sign)
	log.Println("BROADCASTING TRANSACTION TO OTHER PEERS")
	log.DefaultPrint("===========================================")
	BroadcastTransaction(msg,cli.Peers)
}

func (cli *Client)handleConnection(conn net.Conn,url string) {
	for {
		buf1 := make([]byte,4)
		_,err := conn.Read(buf1)
		l := binary.LittleEndian.Uint32(buf1[:4])
		buf := make([]byte,l)
		_,err = conn.Read(buf)
		if(err != nil) {
			break;
		}
		var data sensordata
		json.Unmarshal(buf,&data)
		log.Println("RECIEVED DATA FROM SMARTMETER - "+data.SmID)
		fmt.Println("		SensorData -",data.Data)
		fmt.Println("		SensorName -",data.SensorName)
		fmt.Println("		Timestamp -",data.Start)
		log.DefaultPrint("==========================================")
		cli.createTransaction(buf,url)
	}
	defer conn.Close()
}


// RecieveSensorData is a method called in the main
func (cli *Client)RecieveSensorData(url string) {
	consensus.SnapshotInterval = 1000
	p := cli.Peers
	for {
		p.Mux.Lock()
		l := len(p.Fds)
		p.Mux.Unlock()
		if l > 0 {
			break
		}
		time.Sleep(2*time.Second)
	}
	listener,_ := net.Listen("tcp",":8050")
	for {
		conn,_ := listener.Accept()
		go cli.handleConnection(conn,url)
	}
}

/*
package client

import (
	"encoding/json"
	dt "GO-DAG/DataTypes"
	"time"
	"net/http"
	"github.com/google/uuid"
	"GO-DAG/Crypto"
	"GO-DAG/serialize"
	"GO-DAG/consensus"
	"GO-DAG/storage"
	"crypto/ecdsa"
	"log"
	"bytes"
)

type sensordata struct {
	Data string
	SensorName string
	//SensorID string
	SmID string
	Start int64
	//End int64
}

type postRequest struct {
	Meter sensordata
	TxID string
}

// GenerateMessage creates a byte slice to be sent over a network
// for a particular transaction.
func GenerateMessage(b []byte, signature []byte) []byte {
	var l uint32
	l = uint32(len(b))
	var magicNumber uint32
	magicNumber = 1
	payload := append(b,signature...)
	payload = append(serialize.EncodeToBytes(l),payload...)
	payload = append(serialize.EncodeToBytes(magicNumber),payload...)
	return payload
}

// BroadcastTransaction sends the transaction to all the peers
func BroadcastTransaction(b []byte, p *dt.Peers) {
	p.Mux.Lock()
	for _,conn := range p.Fds {
		conn.Write(b)
	}
	p.Mux.Unlock()
}

// GenerateSignature uses Crypto pkg to genearte a signature.
// It is a wrapper aroung Crypto.Sign function
func GenerateSignature(b []byte, PrivateKey *ecdsa.PrivateKey) []byte {
	hash := Crypto.Hash(b)
	signature := Crypto.Sign(hash[:],PrivateKey)
	return signature
}

// Copy generates the Deep copy of the DAG
func Copy(dag *dt.DAG) *dt.DAG {
	bytes,_ := json.Marshal(dag.Graph)
	var copyGraph map[string] dt.Vertex
	json.Unmarshal(bytes,&copyGraph)
	var copyDag dt.DAG
	copyDag.Graph = copyGraph
	copyDag.Genisis = dag.Genisis
	return &copyDag
}

func fakeSensorData(data *sensordata) {
	data.Start = time.Now().Unix()
}

// SimulateClient is used for testing by sending fake data as transactions.
func SimulateClient(p *dt.Peers, PrivateKey *ecdsa.PrivateKey, dag *dt.DAG, url string) {
	consensus.SnapshotInterval = 1000
	var tx dt.Transaction
	var fakeData sensordata
	fakeData.SensorName = "livingroom-sensor"
	ID := Crypto.Hash(Crypto.SerializePublicKey(&PrivateKey.PublicKey))
	//fakeData.SensorID = string(ID[:])
	fakeData.SmID = Crypto.EncodeToHex(ID[:])
	fakeData.Data = "lighton"

	for {
		p.Mux.Lock()
		l := len(p.Fds)
		p.Mux.Unlock()
		if l > 0 {
			break
		}
		time.Sleep(2*time.Second)
	}

	log.Println("Generating Transactions")

	for {
		fakeSensorData(&fakeData)
		tx.Timestamp = time.Now().Unix()
		s,_ := json.Marshal(fakeData)
		tx.Hash = Crypto.Hash(s)
		tx.TxID = uuid.New()
		var request postRequest 
		request.Meter = fakeData
		request.TxID = Crypto.EncodeToHex(tx.TxID[:])
		serial,_ := json.Marshal(request)
		b := bytes.NewReader(serial)
		http.Post(url,"application/json",b)
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
*/