package consensus

import (
	"fmt"
	dt "GO-DAG/DataTypes"
	"math/rand"
	"math"
	"encoding/hex"
	// "GO-DAG/storage"
)

var SnapshotInterval int

func getMax(s []int) int {    
	//Get maximum value from a slice of integers
	var maximum int
	for index, element := range(s) {
		if index == 0 || element > maximum {
			maximum = element
		}
	}
	return maximum
}

func getMin(s []int) int {     
	//Get minimum value from a slice of integers
	var minimum int
	for index, element := range(s) {
		if index == 0 || element > minimum {
			minimum = element
		}
	}
	return minimum
}

func sum(list []float64) float64 {  
	//Gets sum of all elements of a float64 slice
	var sum float64
	for _,element := range(list) {
		sum = sum + element
	}
	return sum
}

// EncodeToHex is a wrapper around hex function
func EncodeToHex(data []byte) string {
	return hex.EncodeToString(data)
}

func generateRandomNumber(min float64, max float64) float64 {    
	//Generates a random number between given minimum and maximum values
	RandomNumber := rand.Float64()*(max-min) + min
	return float64(RandomNumber)
}

func contains(s []string, e string) bool {  
	//Checks if a element(string) is present in the given slice(of strings)
    for _, element := range s {
        if element == e {
            return true
        }
    }
    return false
}

func GetKeysValuesFromMap_int(mymap map[string] int) ([]string, []int){   
	//Get the keys and values from a given type of map seperately as slices
	keys := make([]string, 0, len(mymap))
	values := make([]int, 0, len(mymap))
	for k, v := range mymap {
		keys = append(keys, k)
		values = append(values,v)
	}
	return keys,values
}

func GetKeysValuesFromMap_float64(mymap map[string] float64) ([]string, []float64){   
	//Get the keys and values from a given type of map seperately as slices
	keys := make([]string, 0, len(mymap))              //Map[string] float64
	values := make([]float64, 0, len(mymap))
	for k, v := range mymap {
		keys = append(keys, k)
		values = append(values,v)
	}
	return keys,values
}

func IsTip(Ledger dt.DAG, Transaction string) bool {     
	//Checks if the Given transaction is a tip or not
	if len(Ledger.Graph[Transaction].Neighbours) < 1 {  
		//If the neighbours slice is empty then the tx is tip
		return true
	} else {
		return false
	}
}

func SelectSubgraph(Ledger dt.DAG, LatestMilestone string) []string {  
	//Returns the Subgraph transactions hashes as a slice
	SelectedSubGraph,_ := GetFutureSet_new(Ledger,LatestMilestone)
	// _,c2 := GetFutureSet_new(Ledger,LatestMilestone)
	return SelectedSubGraph
}

func GetFutureSet_new(dag dt.DAG,Transaction string) ([]string,int) {
	// BFS of the DAG to get the future set
	var queue []string
	queue = append(queue,Transaction)
	count := 0
	for {
		if count == len(queue) {
			break
		}
		tx := dag.Graph[queue[count]]
		for _,v := range tx.Neighbours {
			if !contains(queue,v) {
				queue = append(queue,v)
			}
		}
		count = count+1
	}
	return queue,count 
}

func CalculateRating(Ledger dt.DAG,LatestMilestone string) map[string] int {   
	//Calculetes the rating of all the transactions in the subgraph
	SubGraph := SelectSubgraph(Ledger,LatestMilestone)
	//fmt.Println(SubGraph)
	Rating := make(map[string] int)
	for _,transaction := range(SubGraph) {
		_,LengthFutureSet := GetFutureSet_new(Ledger,transaction)
		Rating[transaction] = LengthFutureSet + 1       
		//Rating is the length of the future set of the transaction + 1
	}
	return Rating                                       
	//Returns the Rating map of the hash to the integer value
}

func RatingtoWeights(Rating map[string] int, alpha float64) map[string] float64 { 
	// Gets the weights from the ratings incliding randomness from alpha value
	Weights := make(map[string] float64)
	_, Values := GetKeysValuesFromMap_int(Rating)
	HighestRating := getMax(Values)
	for Hash, Value := range(Rating) {
		NormalisedRating := Value - HighestRating
		Weights[Hash] = math.Exp(float64(NormalisedRating)*alpha)
	}
	return Weights                  
	//Returns the weights a map of hashes of transaction to the float64 weight value
}

func GetEntryPoint(Tips []string) string {
	var EntryPoint string
	var Perm []int
	Perm = rand.Perm(len(Tips))
	EntryPoint = Tips[Perm[0]]
	return EntryPoint
}

func BackTrack(Ratings map[string] int, Threshhold int, dag dt.DAG, startingPoint string) string{
	current := startingPoint
	for {
		if(Ratings[current] > Threshhold || current == dag.Genisis) {
			break
		}
		tx := dag.Graph[current].Tx
		if _,ok := dag.Graph[EncodeToHex(tx.LeftTip[:])] ; !ok {
			break
		}
		current = EncodeToHex(tx.LeftTip[:])
	}
	return current;
} 

func NextStep(Ledger dt.DAG, Transaction string, Weights map[string] float64) string {  
	//Returns the hash of transaction of the next step to be taken in random walk
	var values []float64
	for _,neighbour := range(Ledger.Graph[Transaction].Neighbours) {
		values = append(values,Weights[neighbour])
	}
	RandNumber := generateRandomNumber(0,sum(values))
	var NextTx string
	for _,approver := range(Ledger.Graph[Transaction].Neighbours) {
		RandNumber = RandNumber - Weights[approver]    
		//Next transaction is the one which makes the random number value to zero with its weight substracted
		if RandNumber <= 0 {
			NextTx = approver
			break
		}
	}
	return NextTx
}

func RandomWalk(Ledger dt.DAG, LatestMilestone string, Weights map[string]float64) string {  
	//Returns the tip when given a milestone transaction
	CurrentTransaction := LatestMilestone
	for ;!IsTip(Ledger,CurrentTransaction); {
		NextTransaction := NextStep(Ledger,CurrentTransaction,Weights)
		CurrentTransaction = NextTransaction
	}
	return CurrentTransaction
}

func GetAllTips (Graph map[string] dt.Vertex ) []string {
	var Tips []string
	for k,v := range Graph {
		if len(v.Neighbours) < 1 {
			Tips = append(Tips,k)
		} 
	}
	return Tips
}

func GetTip(Ledger *dt.DAG, alpha float64) string {
	Rating := CalculateRating(*Ledger,Ledger.Genisis)
	Weights := RatingtoWeights(Rating,alpha)
	start := BackTrack(Rating,100,*Ledger,GetEntryPoint(GetAllTips(Ledger.Graph)))
	Tip := RandomWalk(*Ledger,start,Weights)
	/*
	if Rating[Ledger.Genisis] > SnapshotInterval {
		PruneDag(Ledger,Rating)
	}
	*/
	return Tip 
}

// PruneDag prunes the Old Transactions to reduce memory overhead
func PruneDag(dag *dt.DAG, Ratings map[string] int) {
	// select a milestone which references all the tips
	// The subgraph generated by new milestone becomes the new dag
	/*
	if(len(storage.OrphanedTransactions) != 0){
		SnapshotInterval += 100
		return
	}
	*/
	NewMilestone := GetNewMilestone(Ratings,*dag)
	fmt.Println(len(GetAllTips(dag.Graph)))
	//fmt.Println(len(GetAllTips(GetSubGraph(*dag,NewMilestone))),len(GetAllTips(dag.Graph)))
	DeleteOldtransactions(dag,NewMilestone)
	fmt.Println(len(GetAllTips(dag.Graph)))
	dag.Genisis = NewMilestone
	var tip [32]byte
	genesisNode := dag.Graph[dag.Genisis]
	genesisNode.Tx.LeftTip = tip
	genesisNode.Tx.RightTip = tip
	dag.Graph[dag.Genisis] = genesisNode
	fmt.Println(Ratings[dag.Genisis],len(dag.Graph))
	return
}


// DeleteOldtransactions deletes the transactions referenced by NewMilestone
func DeleteOldtransactions(dag *dt.DAG,NewMilestone string) {
	tx := dag.Graph[NewMilestone].Tx
	var tip [32]byte
	if tx.LeftTip == tip && tx.RightTip == tip {
		return 
	}
	l := EncodeToHex(tx.LeftTip[:])
	r := EncodeToHex(tx.RightTip[:])
	if l == r {
		if _,ok := dag.Graph[l]; ok {
		DeleteOldtransactions(dag,l)
		delete(dag.Graph,l)
		}
	} else {
		if _,ok := dag.Graph[l]; ok{
			DeleteOldtransactions(dag,l)
			delete(dag.Graph,l)
		}
		if _,ok := dag.Graph[r]; ok {
			DeleteOldtransactions(dag,r)
			delete(dag.Graph,r)
		}
	}
	return 
}

// GetSubGraph identifies the subGraph formed by Milestone
func GetSubGraph(dag dt.DAG, Milestone string) map[string] dt.Vertex {
	SubDag := make(map[string] dt.Vertex)

	s := SelectSubgraph(dag,Milestone)

	for _,v := range s {
		SubDag[v] = dag.Graph[v]
	}

	return SubDag 
}

func getMaxSet(Ratings map[string] int, neighbours []string) string {
	step := neighbours[0]
	max := Ratings[neighbours[0]]
	for _,v := range neighbours {
		if Ratings[v] > max {
			max = Ratings[v]
			step = v
		}
	}
	return step
}



func GetNewMilestone(Ratings map[string] int, dag dt.DAG) string {
	prev := dag.Genisis 
	path := dag.Genisis

	for {
		neighbours := dag.Graph[path].Neighbours
		path = getMaxSet(Ratings,neighbours)
		if Ratings[path] < SnapshotInterval/2 {
			break
		}
		prev = path
	}
	return prev
}