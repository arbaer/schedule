package eval


import (
	"log"

	"types"
)



var SizeBased = false

func Score_res(jobs []types.JobNode, res []int) (cost int) {
	if SizeBased {
		return Cuts_mincut_maxbw_size(jobs, res)
	}
	return Cuts_mincut_maxbw(jobs, res)

}


func Cuts_mincut_maxbw(jobs []types.JobNode, res []int) (cost int) {
	cost = 0
	totalLen := len(res)

	resLookup := make(map[int]int, totalLen)
	for i,id := range res {
		resLookup[id] = i
	}

	for pos,id := range res {
		stepCost := 0
		job := jobs[id]
		maxOutPos := 0
		edgesDone := 0
		for _,id := range job.OutIds {
			outPos, exists := resLookup[id]
			if exists {
				edgesDone ++
				if maxOutPos < outPos {
					maxOutPos = outPos
				}
			}
		}
		if len(job.OutIds) == 0 {
			//do nothing
		} else if edgesDone == len(job.OutIds) {
			stepCost = maxOutPos - pos
		} else {
			stepCost = totalLen - pos
		}
		cost += stepCost
	}
	return cost
}

func Cuts_mincut_maxbw_size(jobs []types.JobNode, res []int) (cost int) {

//	fmt.Printf("\n")
	cost = 0
	totalLen := len(res)

	resLookup := make(map[int]int, totalLen)
	for i,id := range res {
		resLookup[id] = i
	}

	for pos,id := range res {
		stepCost := 0
		job := jobs[id]
		maxOutPos := 0
		edgesDone := 0
		for _,id := range job.OutIds {
			outPos, exists := resLookup[id]
			if exists {
				edgesDone ++
				if maxOutPos < outPos {
					maxOutPos = outPos
				}
			}
		}
		if len(job.OutIds) == 0 {
			//do nothing
		} else if edgesDone == len(job.OutIds) {
			stepCost = maxOutPos - pos
		} else {
			stepCost = totalLen - pos
		}
		//fmt.Printf("%v * %v (d%v e%v)\n", int(jobs[id].ioTime), stepCost, edgesDone, len(job.OutIds))
		cost += int(jobs[id].IOTime) * stepCost
	}
//	fmt.Printf(" = %v\n\n", cost)
	return cost
}

func DLAResult(jobs []types.JobNode, result []int, show bool) (res float64) {
	return evalDLAResultStep(jobs, result, show, len(result))
}

func evalDLAResultStep(jobs []types.JobNode, result []int, show bool, step int) (res float64) {
	touchedNodes := make(map[int]bool, len(jobs))

//	removeNotUsed = true
	posLookup := make(map[int]int, len(result))
	for pos, nodeId := range result {
		if pos >= step { break; }
		posLookup[nodeId] = pos
	}

	res = 0.0

	for pos, nodeId := range result {
		job := jobs[nodeId]
		if pos >= step { break; }
		//check if all inputs have been processed
		for _, inpId := range job.InIds {
			if _, ok := touchedNodes[inpId]; !ok {
				log.Fatalf("result wrong: %v\ninput %v was not touched before.\n", result, inpId)
			}
		}
		touchedNodes[nodeId] = true


		for _, outId := range job.OutIds {
			res += float64(posLookup[outId] - pos)
		}

	}
	return res
}

func Cuts(jobs []types.JobNode, res []int) (cost int) {
	cost = 0
	cut := 0
	for _,id := range res {
		job := jobs[id]
		cut = cut - len(job.InIds) + len(job.OutIds)
		cost += cut
	}
	return cost
}
