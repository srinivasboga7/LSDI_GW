package gateway

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	// mysql package
	_ "github.com/go-sql-driver/mysql"
)

// handle post requests from clients.

// ...
const (
	MaxBufferSize   = 5
	GWAddress       = "127.0.0.1:8989"
	DatabaseAddress = "172.18.0.1:3306"
)

type sensorData struct {
	CrateID      string
	RecordTime   string
	ReadingValue string
	SensorType   uint8
}

// RunAPI handles all the incoming data from the sensors
func RunAPI() {

	var DataBuffers [4][]sensorData

	http.HandleFunc("/postData", func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		var d sensorData
		err := decoder.Decode(&d)

		if err != nil {
			log.Fatal(err)
		}

		log.Println("DATA RECIEVED FROM SENSOR :", d)

		handleData(&DataBuffers, d)

		return
	})

	http.ListenAndServe(":7000", nil)
}

func handleData(DataBuffers *[4][]sensorData, Data sensorData) {

	sensorType := int(Data.SensorType)

	buffer := append(DataBuffers[sensorType], Data)
	DataBuffers[sensorType] = buffer

	if len(DataBuffers[sensorType]) < MaxBufferSize {

	} else {

		log.Println("BUFFER LIMIT REACHED")

		var txid string

		ser, _ := json.Marshal(DataBuffers[sensorType])
		h := sha256.Sum256(ser)
		hash := hex.EncodeToString(h[:])
		log.Println("HASH VALUE OF THE CHUNK", hash)

		txid = uploadToBlockchain(hash)
		log.Println("HASH UPLOADED TO BLOCKCHAIN")

		for _, data := range DataBuffers[sensorType] {
			uploadToCloud(data, txid, sensorType)
		}
		log.Println("CHUNK UPLOADED TO THE CLOUD")

		DataBuffers[sensorType] = nil

	}

}

// uploads the data to the cloud after getting the transaction ID
func uploadToCloud(data sensorData, txid string, sensorType int) {

	db, err := sql.Open("mysql", "root:root@tcp("+DatabaseAddress+")/coldchain")
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("INSERT INTO RecordStore(CrateID, RecordTime, ReadingValue, SensorType ,TransactionID) VALUES(?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(data.CrateID, data.RecordTime, data.ReadingValue, sensorType, txid)
	if err != nil {
		log.Fatal(err)
	}

	return

}

// uploads the data to the blockchain
func uploadToBlockchain(hash string) string {
	url := "http://" + GWAddress + "/api?q="

	response, err := http.Get(url + hash)
	if err != nil {
		log.Fatal(err)
	}

	TxID, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	txid := hex.EncodeToString(TxID)

	return txid

}
