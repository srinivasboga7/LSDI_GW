package snapshot

import(
	dt "GO-DAG/datatypes"
	"math/rand"
	"encoding/hex"
)
 

// GetAllTips returns all the tips in the graph
//Redefinition of same function in consensus.go to avoid dependency
func GetAllTips_S (Graph map[string] dt.Vertex) []string {
	var Tips []string
	for k,v := range Graph {
		if len(v.Neighbours) < 2 {
			Tips = append(Tips,k)
		} 
	}
	return Tips
}


// PruneDag prunes the Old Transactions to reduce memory overhead
func PruneDag(dag *dt.DAG, Ratings map[string] int) {
	// select a milestone which references all the tips
	NewMilestone := getNewMilestone(Ratings,*dag)
	fmt.Println(len(GetAllTips_S(dag.Graph)))
	DeleteOldtransactions(dag,NewMilestone)
	fmt.Println(len(GetAllTips_S(dag.Graph)))
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
	l := hex.EncodeToString(tx.LeftTip[:])
	r := hex.EncodeToString(tx.RightTip[:])
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



func getNewMilestone(Ratings map[string] int, dag dt.DAG) string {
	prev := dag.Genisis 
	path := dag.Genisis
	for {
		neighbours := dag.Graph[path].Neighbours
		path = getMaxSet(Ratings,neighbours)
		if Ratings[path] < 1000 {
			break
		}
		prev = path
	}
	return prev
}