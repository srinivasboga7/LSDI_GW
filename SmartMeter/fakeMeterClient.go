package main

import(
	"net"
	// "math/rand"
	"github.com/google/uuid"
	"encoding/json"
	"encoding/hex"
	"time"
	"os"
	"log"
)

type sensordata struct {
	Data string
	SensorName string
	//SensorID string
	SmID string
	Start int64
	//End int64
}

func main() {
	conn,err := net.Dial("tcp",os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	simulteSmartMeter(conn)
}

func simulteSmartMeter(conn net.Conn) {
	var fakeData sensordata
	fakeData.Data = "lighton"
	fakeData.SensorName = "livingroom-sensor"
	id := uuid.New()
	fakeData.SmID = hex.EncodeToString(id[:])
	for {
		fakeData.Start = time.Now().Unix()
		msg,_ := json.Marshal(fakeData)
		conn.Write(msg)
		time.Sleep(5*time.Second)
	}
}