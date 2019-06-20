package main

import(
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"GO-DAG/DataTypes"
)


func HandleTransaction(w http.ResponseWriter, r *http.Request) {

	b,err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var data DataTypes.Transaction
	err := json.Unmarshal(b ,&data)
	if ValidTransaction(data) {
		AddTransaction(data)
		BroadcastTransaction(b)
		fmt.Println(string(data))
	}

}

func ValidTransaction(t DataTypes.Transaction) {
	
}

func AddTransaction(t DataTypes.Transaction) {

}



func main() {
	http.HandleFunc("/Transaction",HandleTransaction)
	http.ListenAndServe(":9000",nil)
}