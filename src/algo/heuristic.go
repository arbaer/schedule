package algo

import (
	"fmt"
	"math"
//	"math/rand"

	"types"
	"eval"
)

/**
***      A* Algorithm
**/

/*func GetRand(a map[int]bool) int {
	// produce a pseudo-random number between 0 and len(a)-1
	i := int(float32(len(a)) * rand.Float32())
	for k, _ := range a {
		if i == 0 {
			return k
		} else {
			i--
		}
	}
	panic("impossible")
}
*/

func heuristic_depth(jobs []types.JobNode, result []int, visitsInt []int) (int, []int) {
	//find id of "root"

	res := make([]int, len(result))
	copy(res, result)
	step := 0
	nodeId := -1
	visited := make([]bool, len(res))

	if res[0] == -1 {
		step = 0
		for id,job := range jobs {
			if len(job.InIds) == 0 {
				nodeId = id
				break
			}
		}
	} else {
		for i,id := range res {
			if id == -1 {
				step = i-1
				//fmt.Printf("res:%v\n", res)
				nodeId = res[step]
				break
			}
			visited[id] = true
		}
	}
	score := -1
	if nodeId == -1 {
		//score = evalCuts2(jobs, res)
		score = eval.Score_res(jobs, res)
		//fmt.Printf("   score:%v nodeId%v res%v visited%v step%v\n", score, nodeId, res, step)
	} else {
		score = heuristic_depthRec(jobs, nodeId, res, visited, step, visitsInt)
	}
//	fmt.Printf("res:%v score:%v\n", res, score)
//	fmt.Printf("   score:%v nodeId%v res%v step%v\n", score, nodeId, res, step)
	return score, res
}

func heuristic_depthRec(jobs []types.JobNode, nodeId int, res []int, vis []bool, step int, visitsInt []int) (int) {
	node := jobs[nodeId]
	res[step] = node.Id
	vis[nodeId] = true

	if step == len(jobs)-1 {
		return eval.Score_res(jobs, res)
	}

	cands := getCands2(jobs, vis)

	maxDepth := 0
	for id,_ := range cands {
		if jobs[id].Depth > maxDepth {
			maxDepth = jobs[id].Depth
		}
	}
	//fmt.Printf("cands:%v md: %v\n", cands, maxDepth)
	for id,_ := range cands {
		if jobs[id].Depth < maxDepth {
			delete(cands, id)
		}
	}

/*	minVis := math.MaxInt32
	for id,_ := range cands {
		if visitsInt[id] < minVis {
			minVis = visitsInt[id]
		}
	}
	for id,_ := range cands {
		if visitsInt[id] > minVis {
			delete(cands, id)
		}
	}
*/


/*	maxSize := 0.0
	for id,_ := range cands {
		if jobs[id].ioTime > maxSize {
			maxSize = jobs[id].ioTime
		}
	}
	//fmt.Printf("cands:%v md: %v\n", cands, maxDepth)
	for id,_ := range cands {
		if jobs[id].ioTime < maxSize {
			delete(cands, id)
		}
	}
	minSize := math.MaxInt32
	for id,_ := range cands {
		if int(jobs[id].ioTime) < minSize {
			minSize = int(jobs[id].ioTime)
		}
	}
	//fmt.Printf("cands:%v md: %v\n", cands, maxDepth)
	for id,_ := range cands {
		if int(jobs[id].ioTime) > minSize {
			delete(cands, id)
		}
	}


	nextBestScore := math.MaxInt32
	for id, _ := range cands {
		res[step+1] = id
		sc := score_res(jobs, res[:step+2])

		if sc < nextBestScore {
			nextBestScore = sc
		}
	}
	for id, _ := range cands {
		res[step+1] = id
		sc := score_res(jobs, res[:step+2])

		if sc < nextBestScore {
			delete(cands, id)
		}
	}
*/
	
/*
	minId := math.MaxInt32
	for id,_ := range cands {
		minId = id
	}

	minId := math.MaxInt32
	for id,_ := range cands {
		if id < minId {
			minId = id
		}
	}
*/
	minId := math.MaxInt32
	for id,_ := range cands {
		if id < minId {
			minId = id
		}
	}

	return heuristic_depthRec(jobs, minId, res, vis, step+1, visitsInt)
}

func Heuristic(jobs []types.JobNode) []int {
	for id, _ := range jobs {
		jobs[id].Depth = countAncestors(jobs, id, 0)

		//fmt.Printf("(id:%v depth:%v) ", id, jobs[id].Depth)
	}
	//fmt.Printf("\n")

	res := make([]int, len(jobs))
	for i,_ := range res {
		res[i] = -1
	}
	vis := make([]bool, len(jobs))

	visitsInt := make([]int, len(jobs))
	for _, job := range jobs {
		visitRek(job.Id, jobs, &visitsInt)
	}

	cands := getCands2(jobs, vis)
	step := 0
	nodesTotal := 0
	for len(cands) > 0 {
		bestCost := math.MaxInt32
		bestCand := -1
		var bestRes []int

		for c,_ := range cands {
			res[step] = c
			cost, res := heuristic_depth(jobs, res, visitsInt)
			//fmt.Printf("cost:%v res:%v\n", cost, res)
			if cost < bestCost {
				bestCand = c
				bestCost = cost
				bestRes = res
			}
			nodesTotal++
		}
		res[step] = bestCand
		vis[bestCand] = true
		fmt.Printf("res: %v cost: %v\n", bestRes, bestCost)

		//fmt.Printf("res:%v cost:%v\n", res, bestCost)	

		cands = getCands2(jobs, vis)
		step++
	}

	fmt.Printf("best score MAXBW: %v nodesTotal:%v res:%v\n", eval.Score_res(jobs, res), nodesTotal, res)
	return res
}


/**
*** END   A* Algorithm
**/
