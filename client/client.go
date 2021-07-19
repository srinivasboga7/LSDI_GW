package client

import (
	"LSDI_GW/Crypto"
	dt "LSDI_GW/DataTypes"
	pow "LSDI_GW/Pow"
	"LSDI_GW/consensus"
	"LSDI_GW/p2p"
	"LSDI_GW/serialize"
	"LSDI_GW/storage"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"log"
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
	Send       chan p2p.Msg
	DAG        *dt.DAG
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

// IssueTransaction is used for generating transaction given a hash value to be stored in the transaction
// Input : hash value that is stored in a transaction
// Output : Transaction ID of the generated transaction
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

func (cli *Client) SimulateClient() {

	for i := 0; i < 10; i++ {
		hash := Crypto.Hash([]byte(time.Now().String()))
		cli.IssueTransaction(hash[:])
		time.Sleep(time.Second)
	}

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
