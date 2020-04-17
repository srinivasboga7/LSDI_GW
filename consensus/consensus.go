package consensus

import (
	"GO-DAG/Crypto"
	dt "GO-DAG/DataTypes"
	"bytes"
	"encoding/hex"
	"math"
	"math/rand"
	// "GO-DAG/dt"
)

func getMax(s []int) int {
	//Get maximum value from a slice of integers
	var maximum int
	maximum = s[0]
	for _, element := range s {
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
	for _, element := range s {
		if element > minimum {
			minimum = element
		}
	}
	return minimum
}

func sum(list []float64) float64 {
	//Gets sum of all elements of a float64 slice
	var sum float64
	for _, element := range list {
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

func getKeysValuesFromMapInt(mymap map[string]int) ([]string, []int) {
	//Get the keys and values from a given type of map seperately as slices
	keys := make([]string, 0, len(mymap))
	values := make([]int, 0, len(mymap))
	for k, v := range mymap {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func isTip(Graph map[string]dt.Vertex, Transaction string, Threshhold int) bool {
	//Checks if the Given transaction is a tip or not
	if len(Graph[Transaction].Neighbours) < Threshhold {
		//If the neighbours slice is empty then the tx is tip
		return true
	}
	return false
}

func getFutureSet(Graph map[string]dt.Vertex, Transaction string) ([]string, int) {
	// BFS of the DAG to get the future set
	var queue []string
	queue = append(queue, Transaction)
	count := 0
	for {
		if count == len(queue) {
			break
		}
		tx := Graph[queue[count]]
		for _, v := range tx.Neighbours {
			if !contains(queue, v) {
				queue = append(queue, v)
			}
		}
		count = count + 1
	}
	return queue, count
}

func calculateFutureSet(Graph map[string]dt.Vertex, currPoint string, FutureSet map[string]map[string]bool) {
	futureset := make(map[string]bool)

	neighbours := Graph[currPoint].Neighbours

	futureset[currPoint] = true

	if len(neighbours) == 0 {
		FutureSet[currPoint] = futureset
		return
	}

	for _, neighbour := range neighbours {
		if _, ok := FutureSet[neighbour]; !ok {
			calculateFutureSet(Graph, neighbour, FutureSet)
		}
		for n := range FutureSet[neighbour] {
			futureset[n] = true
		}
	}
	FutureSet[currPoint] = futureset
}

func calRating(Graph map[string]dt.Vertex, start string) map[string]int {

	Rating := make(map[string]int)
	FutureSet := make(map[string]map[string]bool)

	calculateFutureSet(Graph, start, FutureSet)

	for k, v := range FutureSet {
		Rating[k] = len(v)
	}

	return Rating
}

func calculateAllRatings(Graph map[string]dt.Vertex) map[string]int {
	Ratings := make(map[string]int)
	var zero [32]byte
	for h, vertex := range Graph {
		if bytes.Compare(vertex.Tx.LeftTip[:], zero[:]) == 0 && bytes.Compare(vertex.Tx.RightTip[:], zero[:]) == 0 {
			FutureSet := make(map[string]map[string]bool)
			calculateFutureSet(Graph, h, FutureSet)
			for k, v := range FutureSet {
				Ratings[k] = len(v)
			}

		}
	}
	return Ratings
}

// SelectSubgraph ...
func SelectSubgraph(Graph map[string]dt.Vertex, LatestMilestone string) []string {
	//Returns the Subgraph transactions hashes as a slice
	SelectedSubGraph, _ := getFutureSet(Graph, LatestMilestone)
	// _,c2 := GetFutureSet_new(Ledger,LatestMilestone)
	return SelectedSubGraph
}

// CalculateRating computes rating for each vertex in the DAG.
func CalculateRating(Graph map[string]dt.Vertex, LatestMilestone string) map[string]int {
	//Calculetes the rating of all the transactions in the subgraph
	SubGraph := SelectSubgraph(Graph, LatestMilestone)
	//fmt.Println(SubGraph)
	Rating := make(map[string]int)
	for _, transaction := range SubGraph {
		_, LengthFutureSet := getFutureSet(Graph, transaction)
		Rating[transaction] = LengthFutureSet + 1
		//Rating is the length of the future set of the transaction + 1
	}
	return Rating
	//Returns the Rating map of the hash to the integer value
}

// RatingtoWeights converts ratings to weights using alpha parameter
func RatingtoWeights(Rating map[string]int, alpha float64) map[string]float64 {
	// Gets the weights from the ratings incliding randomness from alpha value
	Weights := make(map[string]float64)
	_, Values := getKeysValuesFromMapInt(Rating)
	HighestRating := getMax(Values)
	for Hash, Value := range Rating {
		NormalisedRating := Value - HighestRating
		Weights[Hash] = math.Exp(float64(NormalisedRating) * alpha)
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
func BackTrack(Threshhold int, Graph map[string]dt.Vertex, Genisis string, startingPoint string) string {
	current := startingPoint
	for {
		Rating := calRating(Graph, current)
		rating := Rating[current]
		if rating > Threshhold || current == Genisis {
			break
		}
		tx := Graph[current].Tx
		if _, ok := Graph[EncodeToHex(tx.LeftTip[:])]; !ok {
			break
		}
		current = EncodeToHex(tx.LeftTip[:])
	}
	return current
}

// NextStep returns the pointer to the next transaction in random walk
func NextStep(Graph map[string]dt.Vertex, Transaction string, Weights map[string]float64) string {
	//Returns the hash of transaction of the next step to be taken in random walk
	var neighbours []string
	neighbours = make([]string, len(Graph[Transaction].Neighbours))
	copy(neighbours, Graph[Transaction].Neighbours)

	var total float64
	for _, neighbour := range neighbours {
		total += Weights[neighbour]
	}
	RandNumber := generateRandomNumber(0, total)

	var approver string
	for _, approver = range neighbours {
		RandNumber -= Weights[approver]
		//Next transaction is the one which makes the random number value to zero with its weight substracted
		if math.Signbit(RandNumber) {
			break
		}
	}
	return approver
}

// RandomWalk is an Implementation of IOTA's MCMC algorithm
func RandomWalk(Graph map[string]dt.Vertex, LatestMilestone string, Ratings map[string]int, alpha float64) string {
	//Returns the tip when given a milestone transaction
	CurrentTransaction := LatestMilestone
	for {
		rating := make(map[string]int)
		if len(Graph[CurrentTransaction].Neighbours) < 2 {
			break
		}
		for _, neighbour := range Graph[CurrentTransaction].Neighbours {
			rating[neighbour] = Ratings[neighbour]
		}
		Weights := RatingtoWeights(rating, alpha)
		CurrentTransaction = NextStep(Graph, CurrentTransaction, Weights)
	}
	return CurrentTransaction
}

// GetAllTips returns all the tips in the graph
func GetAllTips(Graph map[string]dt.Vertex) []string {
	var Tips []string
	var zero [32]byte
	for k, v := range Graph {
		if len(v.Neighbours) < 2 {
			if bytes.Compare(Crypto.DecodeToBytes(k), zero[:]) != 0 {
				Tips = append(Tips, k)
			}
		}
	}
	return Tips
}

// GetTip returns the tip after a random walk from a point chosen by BackTrack
func GetTip(Ledger *dt.DAG, alpha float64) string {
	Ledger.Mux.Lock()
	// if len(Ledger.Graph) > 1000 {
	// 	allRatings := calculateAllRatings(Ledger.Graph)
	// 	PruneDag(Ledger, allRatings, 1000)
	// }
	start := BackTrack(50, Ledger.Graph, Ledger.Genisis, GetEntryPoint(GetAllTips(Ledger.Graph)))
	Graph := GetSubGraph(Ledger.Graph, start)

	Ledger.Mux.Unlock()

	Rating := calRating(Graph, start)
	Tip := RandomWalk(Graph, start, Rating, alpha)

	var zero [32]byte
	if bytes.Compare(Crypto.DecodeToBytes(Tip), zero[:]) == 0 {
		panic("Null tip selected")
	}
	return Tip
}

// GetSubGraph identifies the subGraph formed by Milestone
func GetSubGraph(Graph map[string]dt.Vertex, Milestone string) map[string]dt.Vertex {
	SubDag := make(map[string]dt.Vertex)
	futureset := make(map[string]map[string]bool)
	calculateFutureSet(Graph, Milestone, futureset)
	s := futureset[Milestone]
	for v := range s {
		var copyVertex dt.Vertex
		copyVertex.Tx = Graph[v].Tx
		copy(copyVertex.Neighbours, Graph[v].Neighbours)
		copyVertex.Signature = Graph[v].Signature
		SubDag[v] = copyVertex
	}
	return SubDag
}

// PruneDag prunes the Old Transactions to reduce memory overhead
func PruneDag(dag *dt.DAG, Ratings map[string]int, Threshold int) {

	for k, v := range dag.Graph {
		if Ratings[k] > Threshold {
			for _, neighbour := range v.Neighbours {
				clearTips(dag.Graph[neighbour], k)
			}
			delete(dag.Graph, k)
		}
	}

	return
}

func clearTips(v dt.Vertex, tip string) {
	var t [32]byte
	if hex.EncodeToString(v.Tx.LeftTip[:]) == tip {
		v.Tx.LeftTip = t
	}
	if hex.EncodeToString(v.Tx.RightTip[:]) == tip {
		v.Tx.RightTip = t
	}
	return
}
