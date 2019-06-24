package server

import(
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"GO-DAG/DataTypes"
)


func HandleTransaction(IPs []string) {


}


func ValidTransaction(t DataTypes.Transaction) bool {

	return true
	
}

/*
func AddTransaction(t DataTypes.Transaction) {

}
*/

func BroadcastTransaction(t []bytes, IPs []string) {


}


func StartServer(NodeIPs []string) {
	
}