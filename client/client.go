package client

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	pow "GO-DAG/Pow"
	"GO-DAG/consensus"
	"GO-DAG/p2p"
	"GO-DAG/serialize"
	"GO-DAG/storage"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Client ...
type Client struct {
	PrivateKey *ecdsa.PrivateKey
	// should I keep the dag here or run a go routine for storage layer
	Send chan p2p.Msg
	DAG  *dt.DAG
}

// TxStats ...
type TxStats struct {
	TxInputRate float64
}

// CalculateTxInputRate function calculates the input transaction rate
func (cli *Client) CalculateTxInputRate(windowSize int64) (float64, int) {
	cli.DAG.Mux.Lock()
	window := time.Now().UnixNano() - windowSize*time.Second.Nanoseconds()
	var txsRate float64
	l := len(cli.DAG.RecentTXs)
	for i := l - 1; i >= 0; i-- {
		if cli.DAG.RecentTXs[i] > window {
			txsRate++
		} else {
			break
		}
	}
	cli.DAG.Mux.Unlock()
	return txsRate / float64(windowSize), l
}

// IssueTransaction ...
func (cli *Client) IssueTransaction(hash []byte) []byte {
	var tx dt.Transaction
	copy(tx.Hash[:], hash[:])
	tx.Timestamp = time.Now().UnixNano()
	copy(tx.From[:], Crypto.SerializePublicKey(&cli.PrivateKey.PublicKey))
	// tip selection
	// broadcast transaction

	copy(tx.LeftTip[:], Crypto.DecodeToBytes(consensus.GetTip(cli.DAG, 0.001)))
	copy(tx.RightTip[:], Crypto.DecodeToBytes(consensus.GetTip(cli.DAG, 0.001)))
	pow.PoW(&tx, 3)
	// fmt.Println("After pow")
	b := serialize.Encode32(tx)
	var msg p2p.Msg
	msg.ID = 32
	h := Crypto.Hash(b)
	sign := Crypto.Sign(h[:], cli.PrivateKey)
	msg.Payload = append(b, sign...)
	msg.LenPayload = uint32(len(msg.Payload))
	cli.Send <- msg
	storage.AddTransaction(cli.DAG, tx, sign)
	return h[:]
}

// SimulateClient issues fake transactions
func (cli *Client) SimulateClient() {

	if triggerServer() {
		time.Sleep(5 * time.Second)
		i := 0
		for {
			fmt.Println(i, "===============")
			hash := Crypto.Hash([]byte("Hello,World!"))
			cli.IssueTransaction(hash[:])
			i++
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func triggerServer() bool {
	listener, err := net.Listen("tcp", ":6666")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := listener.Accept()
	buf := make([]byte, 1)
	conn.Read(buf)
	if buf[0] == 0x05 {
		return true
	}
	return false
}

type query struct {
	hash string
}

// GetMemUsage returns the memory used by the current process
func GetMemUsage() uint64 {
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)
	return mem.Alloc
}

// GetCPUUsage returns the percentage of CPU used by the process
func GetCPUUsage() (string, error) {
	pid := os.Getpid()
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "%cpu").Output()
	if err != nil {
		return "", err
	}
	percent := strings.Split(string(out), "\n")[1]
	return percent, nil
}

// RunAPI implements RESTAPI
func (cli *Client) RunAPI() {

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {

		var hash string
		for _, v := range r.URL.Query() {
			hash = v[0]
		}
		//log.Println(q.hash)
		h, _ := hex.DecodeString(hash)
		TxID := cli.IssueTransaction(h)
		w.WriteHeader(http.StatusOK)
		// may be wrap it in a json object
		w.Write(TxID)
		return
	})

	http.HandleFunc("/CPUStats", func(w http.ResponseWriter, r *http.Request) {
		// get request for CPU stats
		cpu, err := GetCPUUsage()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		cpu = cpu + "percent"
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cpu))
		return
	})

	http.HandleFunc("/MemStats", func(w http.ResponseWriter, r *http.Request) {
		// get request for Mem Stats
		mem := strconv.Itoa(int(GetMemUsage()))
		mem = mem + "bytes"
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mem))
		return
	})

	http.HandleFunc("/TxInputRate", func(w http.ResponseWriter, r *http.Request) {
		txRate, _ := cli.CalculateTxInputRate(30)
		var txStats TxStats
		txStats.TxInputRate = txRate
		w.WriteHeader(http.StatusOK)
		s, _ := json.Marshal(txStats)
		w.Write(s)
		return
	})

	http.HandleFunc("/CurrentTips", func(w http.ResponseWriter, r *http.Request) {
		cli.DAG.Mux.Lock()
		tips := consensus.GetAllTips(cli.DAG.Graph, 1)
		cli.DAG.Mux.Unlock()
		type TipsResp struct {
			tips []string
		}
		var resp TipsResp
		resp.tips = tips

		w.WriteHeader(http.StatusOK)
		s, _ := json.Marshal(resp)
		w.Write(s)
		return
	})

	http.HandleFunc("/CurrentDAGSize", func(w http.ResponseWriter, r *http.Request) {
		cli.DAG.Mux.Lock()
		DAGLength := len(cli.DAG.Graph)
		cli.DAG.Mux.Unlock()
		type DAGSize struct {
			length int
		}
		var resp DAGSize
		resp.length = DAGLength

		w.WriteHeader(http.StatusOK)
		s, _ := json.Marshal(resp)
		w.Write(s)
	})

	http.ListenAndServe(":8989", nil)
}
