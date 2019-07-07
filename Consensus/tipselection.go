package TipSelection

import (
	dt "GO-DAG/DataTypes"
	"math/rand"
	"math"
    "time"
)

func GetMax(s []int) int {    //Get maximum value from a slice of integers
	var maximum int
	for index, element := range(s) {
		if index == 0 || element > maximum {
			maximum = element
		}
	}
	return maximum
}

func GetMin(s []int) int {     //Get minimum value from a slice of integers
	var minimum int
	for index, element := range(s) {
		if index == 0 || element > minimum {
			minimum = element
		}
	}
	return minimum
}

func Sum(list []float64) float64 {  //Gets sum of all elements of a float64 slice
	var sum float64
	for _,element := range(list) {
		sum = sum + element
	}
	return sum
}

func GenerateRandomNumber(min int, max int) float64 {    //Generates a random number between given minimum and maximum values
	rand.Seed(time.Now().UnixNano())
	RandomNumber := rand.Intn(max-min) + min
	return float64(RandomNumber)
}

func contains(s []string, e string) bool {  //Checks if a element(string) is present in the given slice(of strings)
    for _, element := range s {
        if element == e {
            return true
        }
    }
    return false
}

func GetKeysValuesFromMap_int(mymap map[string] int) ([]string, []int){   //Get the keys and values from a given type of map seperately as slices
	keys := make([]string, 0, len(mymap))                                 // Map[string] int
	values := make([]int, 0, len(mymap))
	for k, v := range mymap {
		keys = append(keys, k)
		values = append(values,v)
	}
	return keys,values
}

func GetKeysValuesFromMap_float64(mymap map[string] float64) ([]string, []float64){   //Get the keys and values from a given type of map seperately as slices
	keys := make([]string, 0, len(mymap))              //Map[string] float64
	values := make([]float64, 0, len(mymap))
	for k, v := range mymap {
		keys = append(keys, k)
		values = append(values,v)
	}
	return keys,values
}

func IsTip(Ledger dt.DAG, Transaction string) bool {     //Checks if the Given transaction is a tip or not
	if len(Ledger.Graph[Transaction].Neighbours) == 0 {  //If the neighbours slice is empty then the tx is tip
		return true
	} else {
		return false
	}
}

func SelectSubgraph(Ledger dt.DAG, LatestMilestone string) []string {  //Returns the Subgraph transactions hashes as a slice
	SelectedSubGraph,_ := GetFutureSet(Ledger,LatestMilestone)
	return SelectedSubGraph
}

func GetFutureSet(Ledger dt.DAG,Transaction string) ([]string,int) {  //Gets the set of all approvers direct or indirect of a given transaction
	AllFutureSet := make([]string,0)
	WorkingSlice := make([]string,0)
	WorkingSlice = append(WorkingSlice,Transaction)
	for {
		NewSlice := make([]string,0)
		for _,itx := range(WorkingSlice) {
			for _,ltx := range(Ledger.Graph[itx].Neighbours) {
				if !contains(AllFutureSet,ltx) && !contains(WorkingSlice,ltx) && !contains(NewSlice,ltx) {  //Avoid Duplicates
					NewSlice = append(NewSlice,ltx)
				}
			}
		}
		AllFutureSet = append(AllFutureSet,WorkingSlice...)
		if len(NewSlice) == 0 {
			break
		}
		WorkingSlice := make([]string,len(NewSlice))		
		copy(WorkingSlice,NewSlice)	
	}
	return AllFutureSet,len(AllFutureSet)           //Returns the slice of hashes of the future set and the length of the future set
}

func CalculateRating(Ledger dt.DAG,LatestMilestone string) map[string] int {   //Calculetes the rating of all the transactions in the subgraph
	SubGraph := SelectSubgraph(Ledger,LatestMilestone)
	Rating := make(map[string] int)
	for _,transaction := range(SubGraph) {
		_,LengthFutureSet := GetFutureSet(Ledger,transaction)
		Rating[transaction] = LengthFutureSet + 1       //Rating is the length of the future set of the transaction + 1
	}
	return Rating                                       //Returns the Rating map of the hash to the integer value
}

func RatingtoWeights(Rating map[string] int, alpha float64) map[string] float64 { // Gets the weights from the ratings incliding randomness from alpha value
	Weights := make(map[string] float64)
	_, Values := GetKeysValuesFromMap_int(Rating)
	HighestRating := GetMax(Values)
	for Hash, Value := range(Rating) {
		NormalisedRating := Value - HighestRating
		Weights[Hash] = math.Exp(float64(NormalisedRating)*alpha)
	}
	return Weights                  //Returns the weights a map of hashes of transaction to the float64 weight value
}

func NextStep(Ledger dt.DAG, Transaction string, Weights map[string] float64) string {  //Returns the hash of transaction of the next step to be taken in random walk
	_, values := GetKeysValuesFromMap_float64(Weights)
	RandNumber := GenerateRandomNumber(0,int(Sum(values)))
	var NextTx string
	for _,approver := range(Ledger.Graph[Transaction].Neighbours) {
		RandNumber = RandNumber - Weights[approver]    //Next transaction is the one which makes the random number value to zero with its weight substracted
		if RandNumber <= 0 {
			NextTx = approver
			break
		}
	}
	return NextTx
}

func RandomWalk(Ledger dt.DAG, LatestMilestone string, Weights map[string]float64) string {  //Returns the tip when given a milestone transaction
	CurrentTransaction := LatestMilestone
	for ;!IsTip(Ledger,CurrentTransaction); {
		NextTransaction := NextStep(Ledger,CurrentTransaction, Weights)
		CurrentTransaction = NextTransaction
	}
	return CurrentTransaction
}