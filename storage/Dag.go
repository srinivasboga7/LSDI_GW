package storage

import(
	dt "GO-DAG/DataTypes"
	"fmt"
	"GO-DAG/serialize"
	"GO-DAG/Crypto"
	"GO-DAG/consensus"
)

func AddTransaction(dag dt.DAG,tx dt.Transaction, signature []byte) bool {

	var node dt.Node
	var duplicationCheck bool
	duplicationCheck = false
	s := serialize.SerializeData(tx)
	hash := Crypto.Hash(s)
	h := Crypto.EncodeToHex(hash[:])
	dag.Mux.Lock()
	if _,ok := dag.Graph[h] ; !ok{  // Duplication check
		node.Tx = tx
		node.Signature = signature
		dag.Graph[h] = node  
		l := dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])]
		r := dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])]
		if l.Tx == r.Tx {
			l.Neighbours = append(l.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
		} else {
			l.Neighbours = append(l.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.LeftTip[:])] = l
			r.Neighbours = append(r.Neighbours,h)
			dag.Graph[Crypto.EncodeToHex(tx.RightTip[:])] = r
		}
		duplicationCheck = true
		fmt.Println("Added Transaction", h)
	}
	dag.Mux.Unlock()
	return duplicationCheck
}

func GetAllTips (Graph map[string] dt.Node ) []string {
	var Tips []string
	for k,v := range Graph {
		if len(v.Neighbours) == 0 {
			Tips = append(Tips,k)
		} 
	}
	return Tips
}


func PruneDag(dag *dt.DAG) {
	// select a milestone which references all the tips
	// The subgraph generated by new milestone becomes the new dag

	dag.Mux.Lock()
	Ratings := consensus.CalculateRating(*dag,dag.Genisis)

	NewMilestone := GetNewMilestone(Ratings,*dag)

	fmt.Println(len(GetAllTips(GetSubGraph(*dag,NewMilestone))),len(GetAllTips(dag.Graph)))

	dag.Genisis = NewMilestone
	dag.Graph = GetSubGraph(*dag,NewMilestone)

	dag.Mux.Unlock()
}


func GetSubGraph(dag dt.DAG, Milestone string) map[string] dt.Node {
	SubDag := make(map[string] dt.Node)

	s := consensus.SelectSubgraph(dag,Milestone)

	for _,v := range s {
		SubDag[v] = dag.Graph[v]
	}

	return SubDag 
}

func GetNewMilestone(Ratings map[string] int, dag dt.DAG) string {
	// 
	var path string
	path = dag.Genisis
	var milestone string

	for i := 0 ; i < 4 ; i++ {
		neighbours := dag.Graph[path].Neighbours
		m := Ratings[neighbours[0]]
		for _,v := range neighbours {
			if Ratings[v] > m {
				m = Ratings[v]
				milestone = v
			}
		}
		path = milestone 
	}
	return milestone
}