package consensus

import (
	// "fmt"
	dt "GO-DAG/DataTypes"
	"GO-DAG/snapshot"
	"math/rand"
	"math"
	"encoding/hex"
	// "GO-DAG/dt"
)

func getMax(s []int) int {    
	//Get maximum value from a slice of integers
	var maximum int
	maximum = s[0]
	for _, element := range(s) {
		if element > maximum {
			maximum = element
		}
	}
	return maximum
}

func getMin(s []int) int {     
	//Get minimum value from a slice of integers
	var minimum int
	minimum = s[0]
	for _, element := range(s) {
		if element > minimum {
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
    for _, element := range(s) {
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
	for k, v := range(mymap) {
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

func IsTip(Graph map[string]dt.Vertex, Transaction string) bool {     
	//Checks if the Given transaction is a tip or not
	if len(Graph[Transaction].Neighbours) < 2 {  
		//If the neighbours slice is empty then the tx is tip
		return true
	} else {
		return false
	}
}

func GetFutureSet_new(Graph map[string] dt.Vertex,Transaction string) ([]string,int) {
	// BFS of the DAG to get the future set
	var queue []string
	queue = append(queue,Transaction)
	count := 0
	for {
		if count == len(queue) {
			break
		}
		tx := Graph[queue[count]]
		for _,v := range tx.Neighbours {
			if !contains(queue,v) {
				queue = append(queue,v)
			}
		}
		count = count+1
	}
	return queue,count 
}

func SelectSubgraph(Graph map[string]dt.Vertex, LatestMilestone string) []string {
	//Returns the Subgraph transactions hashes as a slice
	SelectedSubGraph,_ := GetFutureSet_new(Graph,LatestMilestone)
	// _,c2 := GetFutureSet_new(Ledger,LatestMilestone)
	return SelectedSubGraph
}

func CalculateRating(Graph map[string]dt.Vertex,LatestMilestone string) map[string] int {   
	//Calculetes the rating of all the transactions in the subgraph
	SubGraph := SelectSubgraph(Graph,LatestMilestone)
	//fmt.Println(SubGraph)
	Rating := make(map[string] int)
	for _,transaction := range(SubGraph) {
		_,LengthFutureSet := GetFutureSet_new(Graph,transaction)
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

// GetEntryPoint returns a tip from which we perform backtrack
func GetEntryPoint(Tips []string) string {
	var EntryPoint string
	var Perm []int
	Perm = rand.Perm(len(Tips))
	EntryPoint = Tips[Perm[0]]
	return EntryPoint
}

// BackTrack returns a point whose cumilative weight is greater than threshold
func BackTrack(Threshhold int, Graph map[string] dt.Vertex, Genisis string,startingPoint string) string{
	current := startingPoint
	var rating int 
	for {
		_,rating = GetFutureSet_new(Graph,current)
		if(rating > Threshhold || current == Genisis) {
			break
		}
		tx := Graph[current].Tx
		if _,ok := Graph[EncodeToHex(tx.LeftTip[:])] ; !ok {
			break
		}
		current = EncodeToHex(tx.LeftTip[:])
	}
	return current;
} 

// NextStep returns the pointer to the next transaction in random walk
func NextStep(Graph map[string]dt.Vertex , Transaction string, Weights map[string] float64) string {  
	//Returns the hash of transaction of the next step to be taken in random walk
	var values []float64
	for _,neighbour := range(Graph[Transaction].Neighbours) {
		values = append(values,Weights[neighbour])
	}
	RandNumber := generateRandomNumber(0,sum(values))
	var NextTx string
	for _,approver := range(Graph[Transaction].Neighbours) {
		RandNumber = RandNumber - Weights[approver]    
		//Next transaction is the one which makes the random number value to zero with its weight substracted
		if RandNumber <= 0 {
			NextTx = approver
			break
		}
	}
	return NextTx
}


// RandomWalk is an Implementation of IOTA's MCMC algorithm
func RandomWalk(Graph map[string]dt.Vertex, LatestMilestone string, Weights map[string]float64) string {  
	//Returns the tip when given a milestone transaction
	CurrentTransaction := LatestMilestone
	for ;!IsTip(Graph,CurrentTransaction); {
		NextTransaction := NextStep(Graph,CurrentTransaction,Weights)
		CurrentTransaction = NextTransaction
	}
	return CurrentTransaction
}

// GetAllTips returns all the tips in the graph
func GetAllTips (Graph map[string] dt.Vertex) []string {
	var Tips []string
	for k,v := range Graph {
		if len(v.Neighbours) < 2 {
			Tips = append(Tips,k)
		} 
	}
	return Tips
}

// GetTip returns the tip after a random walk from a point chosen by BackTrack
func GetTip(Ledger *dt.DAG, alpha float64) string {
	start := BackTrack(50,Ledger.Graph,Ledger.Genisis,GetEntryPoint(GetAllTips(Ledger.Graph)))
	Rating := CalculateRating(Ledger.Graph,start)
	Weights := RatingtoWeights(Rating,alpha)
	Tip := RandomWalk(Ledger.Graph,start,Weights)
	if Rating[Ledger.Genisis] > 2000 {
		snapshot.PruneDag(Ledger,Rating)
	}
	return Tip 
}

// // PruneDag prunes the Old Transactions to reduce memory overhead
// func PruneDag(dag *dt.DAG, Ratings map[string] int) {
// 	// select a milestone which references all the tips
// 	NewMilestone := getNewMilestone(Ratings,*dag)
// 	fmt.Println(len(GetAllTips(dag.Graph)))
// 	DeleteOldtransactions(dag,NewMilestone)
// 	fmt.Println(len(GetAllTips(dag.Graph)))
// 	dag.Genisis = NewMilestone
// 	var tip [32]byte
// 	genesisNode := dag.Graph[dag.Genisis]
// 	genesisNode.Tx.LeftTip = tip
// 	genesisNode.Tx.RightTip = tip
// 	dag.Graph[dag.Genisis] = genesisNode
// 	fmt.Println(Ratings[dag.Genisis],len(dag.Graph))
// 	return
// }


// // DeleteOldtransactions deletes the transactions referenced by NewMilestone
// func DeleteOldtransactions(dag *dt.DAG,NewMilestone string) {
// 	tx := dag.Graph[NewMilestone].Tx
// 	var tip [32]byte
// 	if tx.LeftTip == tip && tx.RightTip == tip {
// 		return 
// 	}
// 	l := EncodeToHex(tx.LeftTip[:])
// 	r := EncodeToHex(tx.RightTip[:])
// 	if l == r {
// 		if _,ok := dag.Graph[l]; ok {
// 		DeleteOldtransactions(dag,l)
// 		delete(dag.Graph,l)
// 		}
// 	} else {
// 		if _,ok := dag.Graph[l]; ok{
// 			DeleteOldtransactions(dag,l)
// 			delete(dag.Graph,l)
// 		}
// 		if _,ok := dag.Graph[r]; ok {
// 			DeleteOldtransactions(dag,r)
// 			delete(dag.Graph,r)
// 		}
// 	}
// 	return 
// }

// GetSubGraph identifies the subGraph formed by Milestone
func GetSubGraph(Graph map[string]dt.Vertex, Milestone string) map[string] dt.Vertex {
	SubDag := make(map[string] dt.Vertex)
	s := SelectSubgraph(Graph,Milestone)
	for _,v := range s {
		SubDag[v] = Graph[v]
	}
	return SubDag
}

// func getMaxSet(Ratings map[string] int, neighbours []string) string {
// 	step := neighbours[0]
// 	max := Ratings[neighbours[0]]
// 	for _,v := range neighbours {
// 		if Ratings[v] > max {
// 			max = Ratings[v]
// 			step = v
// 		}
// 	}
// 	return step
// }



// func getNewMilestone(Ratings map[string] int, dag dt.DAG) string {
// 	prev := dag.Genisis 
// 	path := dag.Genisis
// 	for {
// 		neighbours := dag.Graph[path].Neighbours
// 		path = getMaxSet(Ratings,neighbours)
// 		if Ratings[path] < 1000 {
// 			break
// 		}
// 		prev = path
// 	}
// 	return prev
// }