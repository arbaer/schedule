package algo

import (
	"os"
	"log"
	"bufio"
	"math"
	"fmt"
	"sort"

	"types"
	"eval"
)

const Debug = false

/**
***      A* Algorithm
**/


/*func evalCuts_mincut_maxbw(jobs []types.JobNode, res []int) (cost int) {
	curCost := 0
	for _,id := range res {
		lenOut := len(jobs[id].outIds)
		lenIn  := len(jobs[id].inIds)
		curCost = curCost - lenIn + lenOut
		cost += curCost
	}
	return cost
}
*/
//node used in a_star algorithms
type pNode struct {
//	id int
	plotId int
	score int
	g, h int
	result []int
	cands map[int]bool
	pred *pNode
	pred_edge_id int
}

func (p pNode) hash() string {
	arr := make([]int, len(p.result))
	copy(arr, p.result)
	sort.Ints(arr)
	return fmt.Sprintf("%v", arr)
}


func NewpNode_mincut_maxbw() (p pNode) {
	p.cands = make(map[int]bool)
	return p
}


func (p pNode) Step_mincut_maxbw(jobs []types.JobNode, nid int) (r pNode) {
	r = NewpNode_mincut_maxbw()

	vis := make([]bool, len(jobs))
	vis[nid] = true
	pred := p
//	r.g = 0

	resLookup := make(map[int]bool, len(r.result))
	for pred.pred != nil {
		vis[pred.pred_edge_id] = true
	//	r.g += pred.score
		//r.score += (-len(jobs[pred.pred_edge_id].inIds) + len(jobs[pred.pred_edge_id].outIds))
		resLookup[pred.pred_edge_id] = true
//		r.h -= mcuts[pred.pred_edge_id]
		pred = *pred.pred


	}
	r.result = getRes(jobs, p)
	r.g = eval.Score_res(jobs, r.result)

	r.result = append(r.result, nid)
	r.h = heuristic(jobs, r.result)
//	fmt.Printf("heuristic: nid:%v r.h:%v res:%v\n", nid, r.h, res)
	//r.h = p.h - mcuts[nid]
	r.cands = getCands(jobs, vis, jobs[nid])
//	//fmt.Printf("cands:%v vis:%v hash:%v\n", r.cands, vis, r.hash())
	//r.came_from = &p
	return r
}

/*
func heuristic(jobs []types.JobNode, res []int) (h int) {

	return len(jobs)-len(res)
	//return 0
}
*/
func heuristic(jobs []types.JobNode, res []int) (h int) {
	resLookup := make(map[int]bool)

	for _,id := range res {
		resLookup[id] = true
	}

	for id, job := range jobs {
		_, exists := resLookup[id]
		if !exists {
			if len(job.OutIds) > 0 {
				h += len(job.OutIds)
			} else {
				h +=0
			}
		}
	}
	return h
}



func A_star_mincut_maxbw(jobs []types.JobNode) []int {
	for id, _ := range jobs {
		jobs[id].Depth = countAncestors(jobs, id, 0)

		//fmt.Printf("(id:%v depth:%v) ", id, jobs[id].depth)
	}
	//fmt.Printf("\n")

//	pn := NewpNode(len(jobs))
	pn := NewpNode_mincut_maxbw()
	for id, job := range jobs {
		if len(job.InIds) == 0 {
			pn.cands[id] = true
		}
	}

//	mcuts := GetMaxFlows(jobs)
	//fmt.Printf("mcuts:%v\n", mcuts)

	pn.score = 0
	pn.g = 0
//	for i := 0; i < len(jobs); i++ {
//		pn.h += mcuts[i]
//	}
	pn.h = heuristic(jobs, []int{})
	//pn.id = -1

	return a_star_mincut_maxbwIter(jobs, pn)
}

func a_star_mincut_maxbwIter(jobs []types.JobNode, start pNode) []int {
	f, err := os.Create("part_graph_a_star_mincut_maxbw.gv")
	if err != nil {
		log.Fatalf("ERROR:%v", err)
	}
	w := bufio.NewWriter(f)
	w.WriteString(fmt.Sprintf("digraph partition_search_graph {\n"))
	defer f.Close()
	defer w.Flush()
	defer w.WriteString("}")

	closed_set := make(map[string]bool)
	openset := &LookupQueue{}
	openset.Init()
	nodes := make(map[string]*pNode)

	openset.pushQueue(start, start.h)

	epsilon := 0.0//float64(len(jobs))/20.0

	nodesChecked := 0
	nodesPushed := 0
	nodesTotal := 0

//	//fmt.Printf("score:%v \n", f_score[s_hash])
	for openset.Len()  > 0 {
		current, score := openset.popQueue()
		nodesTotal++
		nodesChecked++
		w.WriteString(fmt.Sprintf("\n%v [label=\"c%v,%v\"];\n", current.plotId, current.hash(), score))

		//fmt.Printf("current:%v(\t%v,\t%v,\t%v) h:%v %v\n", score, current.score, current.g, current.h, current.hash(), getRes(jobs, current))
		if len(current.cands) == 0 {
			res := getRes(jobs, current)
			colorRes(jobs, current, w)

			fmt.Printf("cScore:%v (DLA:%v Cuts:%v Cuts_maxbw:%v) best result:%v nodes:%v (p%v t%v)\n", evalResult(jobs, res, false),
				eval.DLAResult(jobs, res, false),
				eval.Cuts(jobs, res),
				eval.Score_res(jobs, res), res, nodesChecked, nodesPushed, nodesTotal)
			for i := 0; i <= len(res); i++ {
				g := eval.Score_res(jobs, res[:i])
				h := heuristic(jobs, res[:i])
				fmt.Printf("score:%v g:%v (h:%v) %v\n", g+h,g ,h, res[:i])
			}
			return res
		}
		closed_set[current.hash()] = true

		for neig,_ := range current.cands {
			nP := current.Step_mincut_maxbw(jobs, neig)

			nodesTotal++
			if nodesTotal % 10000 == 0 {
				log.Printf("nodesTotal:%v nP:%v\n", nodesTotal, nP)
			}
			nP.plotId = nodesTotal

			//fmt.Printf("  neig:(\t%v,\t%v,\t%v) h:%v r:%v\n", nP.score, nP.g, nP.h, nP.hash(), append(getRes(jobs, current), neig))
//			if closed_set[nP.hash()] {
//				continue
//			}
			var old_nP pNode
			if _, exists := nodes[nP.hash()]; exists {
				old_nP = *nodes[nP.hash()]
				//fmt.Printf(" Uneig:(\t%v,\t%v,\t%v) h:%v r:%v\n", old_nP.score, old_nP.g, old_nP.h, old_nP.hash(), getRes(jobs, old_nP))
				nP.plotId = old_nP.plotId
			} else {
				nodes[nP.hash()] = &nP
			}


			res := getRes(jobs, current)
			res = append(res, neig)
			t_g_score := eval.Score_res(jobs, res)
//			t_g_score := nP.score + current.g
			//fmt.Printf("   e:%v t_g_score:%v old_nP.g:%v res:%v\n", !openset.existsQueue(nP), t_g_score, old_nP.g, res)

			neig_f_score := 0
			if !openset.existsQueue(nP) || t_g_score < old_nP.g {//g_score[nP.hash()]  {
				nodesPushed++
				_ = math.MaxInt32
//				neig_f_score = t_g_score + int((1.0 + epsilon)*math.Sqrt(float64(nP.h)))
				neig_f_score = t_g_score + int((1.0 + epsilon)*float64(nP.h))
//				g_score[nP.hash()] = t_g_score
//				neig_f_score = t_g_score + a_star_estimate_cost(jobs, nP)
				nP.pred = &current
				nP.pred_edge_id = neig
//				delete(nodes, nP.hash())
//				nodes[nP.hash()]= &nP
				debug("current", current)
				nP.g = t_g_score
//				fmt.Printf("   exists:%v %v\n", openset.existsQueue(nP), nP.hash())
				if !openset.existsQueue(nP) {
					openset.pushQueue(nP, neig_f_score)
					//fmt.Printf("  pushed score:%v neig:%v (h:%v) r:%v\n", neig_f_score, neig, nP.hash(), getRes(jobs, nP))
				} else {
					openset.updateQueue(nP, neig_f_score)
					//fmt.Printf("  update score:%v neig:%v (h:%v) r:%v\n", neig_f_score, neig, nP.hash(), getRes(jobs, nP))
				}
			}
			if neig_f_score > 0 {
				w.WriteString(fmt.Sprintf("%v [label=\"%v,%v\"];\n", nP.plotId, nP.hash(), neig_f_score))
				w.WriteString(fmt.Sprintf("%v -> %v [label=\"%v(s%v)\"];\n", current.plotId, nP.plotId, neig, neig_f_score))
			}

		}
	}
	return []int{}
}

func getRes(jobs []types.JobNode, current pNode) []int {
	res := make([]int, len(jobs))
	i := len(jobs)

	debug("",current)
	pred := &current
	for pred.pred != nil {
		i--
		res[i] = pred.pred_edge_id
		debug("goal:", i, pred.pred_edge_id)
		pred = pred.pred
	}
	step := i
	for j :=0; j < len(jobs); j++ {
		if j+step < len(jobs) {
			res[j] = res[j+step]
		}
	}
	res = res[:len(jobs)-step]
	return res
}

func colorRes(jobs []types.JobNode, current pNode, w *bufio.Writer) {
	pred := &current
	i := len(jobs)
	for pred.pred != nil {
		i--
		w.WriteString(fmt.Sprintf("%v -> %v [color=\"blue\"];\n", pred.pred.plotId, pred.plotId))
		pred = pred.pred
	}
}

func debug(msg string, args ...interface{}) {
	if !Debug {return}
	fmt.Printf(msg)
	for i := 0; i< len(args); i++ {
		fmt.Printf("%T:%v ", args[i], args[i])
	}
	fmt.Println()
}


/**
*** END   A* Algorithm
**/
