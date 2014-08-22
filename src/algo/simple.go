package algo


import (
//	"os"
	"sort"
	"log"
	"container/list"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"strconv"

	"types"
	"eval"
)

func plotResult(jobs []types.JobNode, result []int, show bool) {
	totalTime := 0.0
	for _, job := range jobs {
		totalTime += job.CpuTime
	}
	log.Printf("Total CPU Time: %v", totalTime)
	cacheMisses := math.MaxInt32
	score := evalResult(jobs, result, false)
	for cs := 0.0; cacheMisses > 0; cs += 1.0 {
		cacheMisses = plotResultSize(jobs, result, false, cs, score)
	}
}

func plotResultSize(jobs []types.JobNode, result []int, show bool, cacheSize float64, score float64) (int) {
	touchedNodes := make(map[int]bool, len(jobs))
	cache := list.New()
	cache.Init()

	res := 0.0
	cacheRes := 0.0
	cacheMisses := 0

	for _, nodeId := range result {
		//check if all inputs have been processed
		for _, inpId := range jobs[nodeId].InIds {
			if _, ok := touchedNodes[inpId]; !ok {
				log.Fatalf("result wrong: %v\ninput %v was not touched before.\n", result, inpId)
			}
		}

		depth := 0.0
		//cacheRes += jobs[nodeId].CpuTime
		moveToFront := make(map[*list.Element]bool)
		for e := cache.Front(); e != nil; e = e.Next() {
			depth += jobs[e.Value.(int)].IOTime
			for _, inpId := range jobs[nodeId].InIds {
				if e.Value == inpId {
					if depth > cacheSize {
						cacheMisses++
						cacheRes += jobs[inpId].IOTime
					}
					moveToFront[e] = true
					res += depth
				}
			}
		}
		for move,_ := range moveToFront {
			cache.MoveToFront(move)
		}
		cache.PushFront(nodeId)
		touchedNodes[nodeId] = true
	}
	if show {
		fmt.Printf("cacheSize: %v, cacheRes: %v, cacheMisses: %v\n", cacheSize, cacheRes, cacheMisses)
	} else {

		//plotFile.WriteString(fmt.Sprintf("%v\t%v\t%v %v\n", cacheSize, cacheRes, *algo, score))
	}
	return cacheMisses
}

func Score_only(jobs []types.JobNode, result string) []int {
	if result[0] != '[' || result[len(result)-1] != ']' {
		log.Fatalf("result is in the wrong format: %v", result)
	}
	result = result[1:len(result)-1]
	numbers := strings.Split(result, " ")
	res := make([]int, len(numbers))

	for i,v := range numbers {
		s, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal(err)
		}
		res[i] = s
	}

	log.Printf("score of given result:%v (DLA:%v CutsBW:%v)", evalResult(jobs, res, false), eval.DLAResult(jobs, res, false), eval.Cuts_mincut_maxbw(jobs, res))
	return res
}

func Score_result(jobs []types.JobNode, result string) []int {
	if result[0] != '[' || result[len(result)-1] != ']' {
		log.Fatalf("result is in the wrong format: %v should be [0 1 2 ....]", result)
	}
	result = result[1:len(result)-1]
	numbers := strings.Split(result, " ")
	res := make([]int, len(numbers))

	for i,v := range numbers {
		s, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal(err)
		}
		res[i] = s
	}

	log.Printf("score of given result:%v", evalResult(jobs, res, true))
	return res
}

func evalResult(jobs []types.JobNode, result []int, show bool) (res float64) {
	return evalResultStep(jobs, result, show, len(result))
}
func evalResultStep(jobs []types.JobNode, result []int, show bool, step int) (res float64) {
	touchedNodes := make(map[int]bool, len(jobs))
	cache := list.New()
	cache.Init()

//	removeNotUsed = true

	res = 0

	for pos, nodeId := range result {
		job := jobs[nodeId]
		if pos >= step { break; }
		//check if all inputs have been processed
		for _, inpId := range job.InIds {
			if _, ok := touchedNodes[inpId]; !ok {
				log.Fatalf("result wrong: %v\ninput %v was not touched before.\n", result, inpId)
			}
		}

		if len(job.InIds) == 0 {
			res += jobs[nodeId].IOTime
		} else {
			inputs := make([]int, len(job.InIds))
			copy(inputs, job.InIds)

			depth := 0.0
			for e := cache.Front(); e != nil; e = e.Next() {
				cslotIds := e.Value.([]int)
//				fmt.Printf("cslot:%v\n", cslotIds)
				for _,id := range cslotIds {
					depth += jobs[id].IOTime
				}
				for _, inpId := range job.InIds {
					for _,id := range cslotIds {
						if id == inpId {
//							fmt.Printf("id:%v inpId:%v res:%v\n", id, inpId, res)
							res += depth
						}
					}
				}
			}
			delete_nested(cache, inputs)
			cache.PushFront(inputs)
		}
		cache.PushFront([]int{nodeId})
		touchedNodes[nodeId] = true

		if show {
			cStr := fmt.Sprintf("%v", cache.Front().Value)
			for e := cache.Front().Next(); e != nil; e = e.Next() {
				//remove the brackets
				innercStr := ""
				nodes := e.Value.([]int)
				for _,id := range nodes {
					innercStr = fmt.Sprintf("%s %d", innercStr, id)
				}
				if len(nodes) > 1 {
					innercStr = fmt.Sprintf(" [%s]", innercStr[1:])
				}
				cStr = fmt.Sprintf("%s ->%v", cStr, innercStr)
			}
			fmt.Printf("res: %v, node: %v%v, cache: %s\n", res, nodeId, jobs[nodeId].InIds, cStr)
		}
	}
	return res
}

func delete_nested(l *list.List, ids []int) {
	for _,remId := range ids {
		for e := l.Front(); e != nil; e =  e.Next() {
			curIds := e.Value.([]int)
			for _,id := range curIds {
				if id == remId {
					if len(curIds) > 1 {
						newIds := make([]int, len(curIds)-1)
						i := 0
						for _,copyId := range curIds {
							if copyId != remId {
								newIds[i] = copyId
								i++
							}
						}
						e.Value = newIds
					} else {
						l.Remove(e)
						break
					}
				}
			}
		}
	}
}

type score struct {
	score float64
	dlaScore float64
	result []int
}

func initScore() (s score) {
	return score{math.MaxFloat64, math.MaxFloat64, []int{}}
}


func CumSubNodes(jobs []types.JobNode) {
	//mark DAG
	visits := make([]int, len(jobs))
	for _, job := range jobs {
		visitRek(job.Id, jobs, &visits)
	}
	log.Printf("visits: %v", visits)

	cands := make(map[int]bool, len(jobs))

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}
	resPointer := 0
	res := make([]int, len(jobs))
	visited := make([]bool, len(jobs))

	for len(cands) > 0 {
		idMinVis := -1
		minVis := math.MaxInt32
		for id, _ := range cands {
			if minVis > visits[id] {
				minVis = visits[id]
				idMinVis = id
			}
		}

		res[resPointer] = idMinVis
		resPointer++
		delete(cands, idMinVis)
		visited[idMinVis] = true
		node := jobs[idMinVis]

		for _, id := range node.OutIds {
			allInVis := true
			for _,inId := range jobs[id].InIds {
				allInVis = allInVis && visited[inId]
			}
			if allInVis && !visited[id] {
				cands[id] = true
			}
		}
		for _, job := range jobs {
			allInVis := true
			for _,inId := range job.InIds {
				allInVis = allInVis && visited[inId]
			}
			if allInVis && !visited[job.Id] {
				cands[job.Id] = true
			}
		}
	}
	log.Printf("cumSubNodes score:%v result: %v\n", evalResult(jobs, res, true), res)
}

func visitRek(id int, jobs []types.JobNode, visits *[]int) {
	//runin from the leafs to the sources
	vis := *visits
	vis[id] = vis[id]+1
	*visits = vis
	for _, inId := range jobs[id].InIds {
		visitRek(inId, jobs, visits)
	}
}

func map2string(x map[int]int) (s string) {
	s = "["
	ar := make([]int, len(x))
	for k,v := range x {
		ar[k] = v
	}
	for i, val := range ar {
		if i > 0 {
			s = fmt.Sprintf("%s, %v:%v", s, i, val-1)

		} else {
			s = fmt.Sprintf("%s%v:%v", s, i,val-1)
		}
	}
	return s+"]"
}

func rpost2array(x map[int]int) (ar []int) {
	ar = make([]int, len(x))
	for k,v := range x {
		ar[v-1] = k
	}
	return ar
}

func DfWalkAll(jobs []types.JobNode) {
	bestScore := initScore()

	cnt := 0
	for ; cnt < len(jobs); cnt++ {
		i := 1
		j := len(jobs)
		pre := make(map[int]int, len(jobs))
		rpost := make(map[int]int, len(jobs))
		nd := make(map[int]int, len(jobs))
		for id, _ := range jobs {
			pre[id] = 0
			rpost[id] = 0
		}

		x := cnt
		for x >= 0 {
			//log.Printf("x %v, pre %v, rpost %v, i %v, j %v, nd %v", x,pre, rpost, i, j, nd)
			pre, rpost, i, j, nd = dfw(x, jobs, pre, rpost, i, j, nd)
			x = -1
			for k,v := range pre {
				if v == 0 {
					x = k
					break
				}
			}

		}
		res := rpost2array(rpost)
		resScore := evalResult(jobs, res, false)
		if resScore < bestScore.score {
			bestScore.score = resScore
			bestScore.result = res
			plotResult(jobs, res, true)
		}
		log.Printf("Final x:%v dfw: score %v, result %v", cnt,evalResult(jobs, res, false), res)
	}
	evalResult(jobs, bestScore.result, true)
	log.Printf("Best: score:%v, res: %v", bestScore.score, bestScore.result)
}

func DfWalk(jobs []types.JobNode) {
	i := 1
	j := len(jobs)
	pre := make(map[int]int, len(jobs))
	rpost := make(map[int]int, len(jobs))
	nd := make(map[int]int, len(jobs))
	for id, _ := range jobs {
		pre[id] = 0
		rpost[id] = 0
	}

	x := -1
	for k,v := range pre {
		log.Printf("k %v, v %v", k,v)
		if v == 0 {
			x = k
			break
		}
	}
	for x >= 0 {
		//log.Printf("x %v, pre %v, rpost %v, i %v, j %v, nd %v", x,pre, rpost, i, j, nd)
		pre, rpost, i, j, nd = dfw(x, jobs, pre, rpost, i, j, nd)
		x = -1
		for k,v := range pre {
			if v == 0 {
				x = k
				break
			}
		}

	}
	res := rpost2array(rpost)
	log.Printf("Final dfw: score %v, result %v",evalResult(jobs, res, true), res)
}

func dfw(x int, jobs []types.JobNode, pre map[int]int, rpost map[int]int, i int, j int, nd map[int]int) (map[int]int, map[int]int, int, int, map[int]int) {
	//log.Printf("x %v, pre %v, rpost %v, i %v, j %v, nd %v", x, pre, rpost, i, j, nd)
	pre[x] = i
	i = i+1
	nd[x] = 1
	for _, y := range jobs[x].OutIds {
		if pre[y] == 0 {
			//labeling: tree arc
			pre, rpost, i, j, nd = dfw(y, jobs, pre, rpost, i, j, nd)
			nd[x] = nd[x] + nd[y]
		} else if rpost[y] == 0 {
			//labeling forward arc
		} else {
			//labeling cross arc
		}
	}
	rpost[x] = j
	j = j - 1
	return pre, rpost, i, j, nd
}

func GreedyAll(jobs []types.JobNode) (res []int){
	//find id of "root"
	//bestScore := score{math.MaxInt32, math.MaxInt32, []int{}}
	bestScore := initScore()
	count := 0
	cands := make(map[int]bool, len(jobs))

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}
	nextBestScore := math.MaxFloat64
	candScores := make(map[int]float64, len(cands))

	result := make([]int, len(jobs))
	visited := make([]bool, len(jobs))

	for id, _ := range cands {
		result[0] = id
		//sc := evalResultStep(jobs, result, false, 1)
		sc := float64(eval.Score_res(jobs, result))
		//sc := float64(eval.DLAResult(jobs, result, false))
		//log.Printf("id:%v sc:%v res:%v", id, sc, result)
		candScores[id] = sc

		if sc < nextBestScore {
			nextBestScore = sc
		}
	}
	//log.Printf("nextBest:%v", nextBestScore)

	for id, _ := range cands {
		if candScores[id] == nextBestScore {
			resClone := make([]int, len(result))
			copy(resClone, result)
			visClone := make([]bool, len(visited))
			copy(visClone, visited)
			result = greedyAllRec(jobs, id, resClone, visClone, 0, &bestScore, &count)
		}
	}
	log.Printf("Finished after checking %v combinations.", count)
	return result
}

func greedyAllRec(jobs []types.JobNode, nodeId int, result []int, visited []bool, step int, bestScore *score, count *int) ([]int) {
	node := jobs[nodeId]


	result[step] = node.Id
	visited[node.Id] = true

	if step == len(jobs)-1 {
		updateBestScore(jobs, result, step, bestScore, count)
		return bestScore.result
	}
	cands := make(map[int]bool, len(jobs))

	for _, id := range node.OutIds {
		allInVis := true
		for _,inId := range jobs[id].InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[id] {
			cands[id] = true
		}
	}
	for _, job := range jobs {
		allInVis := true
		for _,inId := range job.InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[job.Id] {
			cands[job.Id] = true
		}
	}
	//greedy part
	nextBestScore := math.MaxFloat64
	candScores := make(map[int]float64, len(cands))

	for id, _ := range cands {
		result[step+1] = id
		//sc := evalResultStep(jobs, result, false, step+2)
		sc := float64(eval.Score_res(jobs, result))
		//sc := float64(eval.DLAResult(jobs, result[:step+1], false))
//		log.Printf("id:%v sc:%v res:%v", id, sc, result)
		candScores[id] = sc

		if sc < nextBestScore {
			nextBestScore = sc
		}
	}
//	log.Printf("nextBest:%v", nextBestScore)

	for id, _ := range cands {
		if candScores[id] == nextBestScore {
			resClone := make([]int, len(result))
			copy(resClone, result)
			visClone := make([]bool, len(visited))
			copy(visClone, visited)
			greedyAllRec(jobs, id, resClone, visClone, step+1, bestScore, count)
		}
	}
	return bestScore.result
}

func Greedy(jobs []types.JobNode) []int {
	cands := make(map[int]bool, len(jobs))

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}
	res := make([]int, len(jobs))
	for i,_ := range res {
		res[i] = 0
	}
	visited := make([]bool, len(jobs))

	for i := 0; i < len(jobs); i++ {
		nextBest := -1
		nextBestScore := math.MaxFloat64

		for id, _ := range cands {
			res[i] = id
			//sc := evalResultStep(jobs, res, false, i+1)
			sc := float64(eval.Score_res(jobs, res[:i+1]))
			//sc := float64(eval.DLAResult(jobs, res[:i], false))

			if sc < nextBestScore {
				nextBest = id
				nextBestScore = sc
			}
		}
		res[i] = nextBest
		log.Printf("nextBest:%v sc:%v ordering:%v", nextBest, nextBestScore, res[:i+1])
		

		delete(cands, nextBest)
		visited[nextBest] = true

		for _, id := range jobs[nextBest].OutIds {
			allInVis := true
			for _,inId := range jobs[id].InIds {
				allInVis = allInVis && visited[inId]
			}
			if allInVis && !visited[id] {
				cands[id] = true
			}
		}
		for _, job := range jobs {
			allInVis := true
			for _,inId := range job.InIds {
				allInVis = allInVis && visited[inId]
			}
			if allInVis && !visited[job.Id] {
				cands[job.Id] = true
			}
		}
	}

	fmt.Printf("Greedy best score:%v MAXBW: %v result: %v\n", evalResult(jobs, res, false), eval.Score_res(jobs, res), res)
	return res
}

func RandomBreadthFirst(jobs []types.JobNode) {
	cands := make(map[int]bool, len(jobs))

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}
	resPointer := 0
	res := make([]int, len(jobs))
	visited := make([]bool, len(jobs))

	for len(cands) > 0 {
		for id, _ := range cands {
			res[resPointer] = id
			resPointer++
			delete(cands, id)
			visited[id] = true
			node := jobs[id]

			for _, id := range node.OutIds {
				allInVis := true
				for _,inId := range jobs[id].InIds {
					allInVis = allInVis && visited[inId]
				}
				if allInVis && !visited[id] {
					cands[id] = true
				}
			}
			for _, job := range jobs {
				allInVis := true
				for _,inId := range job.InIds {
					allInVis = allInVis && visited[inId]
				}
				if allInVis && !visited[job.Id] {
					cands[job.Id] = true
				}
			}
		}
	}
	log.Printf("BreadthFirst score:%v result: %v\n", evalResult(jobs, res, true), res)
}

func BreadthFirst(jobs []types.JobNode) {
	queue := list.New()
	queue.Init()

	res := make([]int, len(jobs))

	for id, job := range jobs {
		if len(job.Inputs) == 0 {
			queue.PushBack(id)
		}
	}

	cnt := 0
	visited := make([]bool, len(jobs))

	log.Printf("list: %v", queue)
	for queue.Len() > 0 {
		id := queue.Front().Value.(int)
		queue.Remove(queue.Front())

		res[cnt] = id
		visited[id] = true
		cnt++

		log.Printf("id:%v, res: %v, vis: %v", id, res, visited)

		added := make([]bool, len(jobs))
		for _, outid := range jobs[id].OutIds {
			allInVis := true
			for _,inId := range jobs[outid].InIds {
				allInVis = allInVis && visited[inId]
			}
			if !visited[outid] && allInVis && !added[outid] {
				added[outid] = true
				queue.PushBack(outid)
			}
		}
		id = -1
	}

	log.Printf("BreadthFirst score:%v result: %v\n", evalResult(jobs, res, true), res)
}

func ExhaustiveSearch(jobs []types.JobNode) {
	//find id of "root"
	bestScore := initScore()
	count := 0
	cands := make(map[int]bool, len(jobs))

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}
	for id, _ := range cands {
		res := make([]int, len(jobs))
		vis := make([]bool, len(jobs))

		exSearchRec(jobs, id, res, vis, 0, &bestScore, &count)
	}
	log.Printf("Finished after checking %v combinations.", count)
}

func exSearchRec(jobs []types.JobNode, nodeId int, result []int, visited []bool, step int, bestScore *score, count *int) {
	node := jobs[nodeId]


	result[step] = node.Id
	visited[node.Id] = true

	if step == len(jobs)-1 {

		updateBestScore(jobs, result, step, bestScore, count)
	}

	cands := make(map[int]bool, len(jobs))

	for _, id := range node.OutIds {
		allInVis := true
		for _,inId := range jobs[id].InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[id] {
			cands[id] = true
		}
	}
	for _, job := range jobs {
		allInVis := true
		for _,inId := range job.InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[job.Id] {
			cands[job.Id] = true
		}
	}
	//log.Printf("cands: %v", cands)
	for id, _ := range cands {
		resClone := make([]int, len(result))
		copy(resClone, result)
		visClone := make([]bool, len(visited))
		copy(visClone, visited)
		exSearchRec(jobs, id, resClone, visClone, step+1, bestScore, count)
	}
//	log.Printf("step: %v res:  %v\n", step, result)
}

func BreadthEx(jobs []types.JobNode) {
	//find id of "root"
	bestScore := initScore()
	count := 0
	cands := make(map[int]bool, len(jobs))

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}
	for id, _ := range cands {
		candsClone := make(map[int]bool, len(jobs))
		for k,v := range cands {
			if k != id {
				candsClone[k] = v
			}
		}
		res := make([]int, len(jobs))
		vis := make([]bool, len(jobs))

		breadthExRec(jobs, candsClone, id, res, vis, 0, &bestScore, &count)
	}
	log.Printf("Finished after checking %v combinations.", count)
}

func breadthExRec(jobs []types.JobNode, cands map[int]bool, nodeId int, result []int, visited []bool, step int, bestScore *score, count *int) {
	node := jobs[nodeId]
	result[step] = node.Id
	visited[node.Id] = true

	if step == len(jobs)-1 {
		updateBestScore(jobs, result, step, bestScore, count)
		return
	}

	if len(cands) == 0 {
		cands = make(map[int]bool, len(jobs))

		for _, id := range node.OutIds {
			allInVis := true
			for _,inId := range jobs[id].InIds {
				allInVis = allInVis && visited[inId]
			}
			if allInVis && !visited[id] {
				cands[id] = true
			}
		}
		for _, job := range jobs {
			allInVis := true
			for _,inId := range job.InIds {
				allInVis = allInVis && visited[inId]
			}
			if allInVis && !visited[job.Id] {
				cands[job.Id] = true
			}
		}
	}

	for id, _ := range cands {
		candsClone := make(map[int]bool, len(jobs))
		for k,v := range cands {
			if k != id {
				candsClone[k] = v
			}
		}
		delete(candsClone, id)

		resClone := make([]int, len(result))
		copy(resClone, result)
		visClone := make([]bool, len(visited))
		copy(visClone, visited)
		breadthExRec(jobs, candsClone, id, resClone, visClone, step+1, bestScore, count)
	}
}

func countAncestors(jobs []types.JobNode, id int, depth int) int {

//	fmt.Printf("(id:%v depth:%v)\n", id, jobs[id].Depth)

	if len(jobs[id].InIds) == 0 {
		return depth
	}

	max := 0
	for _,ancId := range jobs[id].InIds {
		ancCnt := countAncestors(jobs, ancId, depth+1)
		if ancCnt > max {
			max = ancCnt
		}
	}
	return max
}

func DepthEx(jobs []types.JobNode) {
	//find id of "root"
	bestScore := initScore()
	count := 0
	cands := make(map[int]bool, len(jobs))

	for id, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
		jobs[id].Depth = countAncestors(jobs, id, 0)
		fmt.Printf("final (id:%v depth:%v) ", id, jobs[id].Depth)
//		fmt.Println("")
	}
	for id, _ := range cands {
		candsClone := make(map[int]bool, len(jobs))
		for k,v := range cands {
			if k != id {
				candsClone[k] = v
			}
		}
		res := make([]int, len(jobs))
		vis := make([]bool, len(jobs))

		depthExRec(jobs, candsClone, id, res, vis, 0, &bestScore, &count)
	}
	log.Printf("Finished after checking %v combinations.", count)
}

func depthExRec(jobs []types.JobNode, cands map[int]bool, nodeId int, result []int, visited []bool, step int, bestScore *score, count *int) {
	node := jobs[nodeId]
	result[step] = node.Id
	visited[node.Id] = true

	if step == len(jobs)-1 {
		updateBestScore(jobs, result, step, bestScore, count)
		/*
		*count = *count+1
		if *count%100000 == 0 {
			log.Printf("%v combinations tested. Best score: %v.", *count, bestScore.score)
		}
		scoreVal := evalResult(jobs, result, false)
		if scoreVal < bestScore.score {
			newBest := score{scoreVal, result}
			fmt.Printf("new best: '%v', %v\n", result, scoreVal)
			_ = evalResult(jobs, result, true)
			plotResult(jobs, result, true)
			*bestScore = newBest
		}
		*/
		return
	}

	cands = make(map[int]bool, len(jobs))

	for _, id := range node.OutIds {
		allInVis := true
		for _,inId := range jobs[id].InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[id] {
			cands[id] = true
		}
	}
	for _, job := range jobs {
		allInVis := true
		for _,inId := range job.InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[job.Id] {
			cands[job.Id] = true
		}
	}

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

	for id, _ := range cands {
		candsClone := make(map[int]bool, len(jobs))
		for k,v := range cands {
			if k != id {
				candsClone[k] = v
			}
		}
		delete(candsClone, id)

		resClone := make([]int, len(result))
		copy(resClone, result)
		visClone := make([]bool, len(visited))
		copy(visClone, visited)
		depthExRec(jobs, candsClone, id, resClone, visClone, step+1, bestScore, count)
	}
}
func updateBestScore(jobs []types.JobNode, result []int, step int, bestScore *score, count *int) {
	updateBestScoreMsg(jobs, result, step, bestScore, count, "")
}

func updateBestScoreMsg(jobs []types.JobNode, result []int, step int, bestScore *score, count *int, msg string) {
	*count = *count+1
	if *count%100000 == 0 {
		log.Printf("%v combinations tested. Best score: %v (DLA:%v).", *count, bestScore.score, bestScore.dlaScore)
	}
	//scoreVal := evalResult(jobs, result, false)
	scoreVal := float64(eval.Score_res(jobs, result))
	//scoreVal := float64(eval.DLAResult(jobs, result, false))
	dlaScore := eval.DLAResult(jobs, result, false)
//	if dlaScore != scoreVal {
//		fmt.Printf("HEUREKA!!1!")
//		os.Exit(+2)
//	}
	if scoreVal < bestScore.score {
		fmt.Printf(msg)
		newBest := score{scoreVal, dlaScore, result}
		fmt.Printf("new best: '%v', (MAXBW:%v), (DLA:%v)\n", result, scoreVal, dlaScore)
		_ = evalResult(jobs, result, false)
		plotResult(jobs, result, true)
		*bestScore = newBest
		if dlaScore > bestScore.dlaScore {
			log.Fatalf("%v!!!%v", dlaScore, bestScore.dlaScore)
		}
	}
}

func getCands(jobs []types.JobNode, visited []bool, node types.JobNode) (cands map[int]bool) {
	cands = make(map[int]bool, len(jobs))

	for _, id := range node.OutIds {
		allInVis := true
		for _,inId := range jobs[id].InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[id] {
			cands[id] = true
		}
	}
	for _, job := range jobs {
		allInVis := true
		for _,inId := range job.InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[job.Id] {
			cands[job.Id] = true
		}
	}
	return cands
}

func getCands2(jobs []types.JobNode, visited []bool) (cands map[int]bool) {
	cands = make(map[int]bool, len(jobs))

	for _, job := range jobs {
		allInVis := true
		for _,inId := range job.InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[job.Id] {
			cands[job.Id] = true
		}
	}
	return cands
}

func visitRek2(id int, jobs []types.JobNode, amount float64, visits *[]float64) {
	//runin from the leafs to the sources
	vis := *visits
	vis[id] = vis[id]+amount
	*visits = vis
	amount = amount/float64(len(jobs[id].InIds))
	for _, inId := range jobs[id].InIds {
		visitRek2(inId, jobs, amount, visits)
	}
}

func Baseline(jobs []types.JobNode) []int {
	res := make([]int, len(jobs))
	vis := make([]bool, len(jobs))
	cands := make(map[int]bool)

	for id, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[id] = true
		}
	}
	c := 0
	for len(cands) > 0 {
		currentLayer := make([]int, len(cands))
		i := 0
		for id,_ := range cands {
			currentLayer[i] = id
			vis[id] = true
			i++
		}
		sort.Ints(currentLayer)
		for _,id := range currentLayer {
			delete(cands, id)
			res[c] = id
			c++
		}
		cands = getCands(jobs, vis, jobs[0])
	}
	fmt.Printf("score:%v (DLA:%v CutsBW:%v) best:%v\n",
		evalResult(jobs, res, false),
		eval.DLAResult(jobs, res, false),
		eval.Cuts_mincut_maxbw(jobs, res),
		res)
	return res
}

func GetRand(a map[int]bool) int {
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

func DepthEx2(jobs []types.JobNode) {
	//find id of "root"
	bestScore := initScore()
	count := 0
	cands := make(map[int]bool, len(jobs))

	visitsInt := make([]int, len(jobs))
	for _, job := range jobs {
		visitRek(job.Id, jobs, &visitsInt)
	}
	visits := make([]float64, len(jobs))
	for k,v := range visitsInt {
		visits[k] = float64(v)
	}
	fmt.Printf("visits:")
	for id, vis := range visits {
		fmt.Printf("%v(%v) ", id, vis)
	}
	fmt.Printf("\n")

	for id, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
		jobs[id].Depth = countAncestors(jobs, id, 0)

		fmt.Printf("(id:%v depth:%v) ", id, jobs[id].Depth)
	}
	fmt.Printf("\n")

	fmt.Printf("before cands:%v\n", cands)
	minVis := math.MaxFloat32
	maxVis := 0.0
	for id, _ := range cands {
		if minVis > visits[id] {
			minVis = visits[id]
		}
		if maxVis < visits[id] {
			maxVis = visits[id]
		}
	}
	delsVis := make(map[int]bool)
	for id,_ := range cands {
		if visits[id] > minVis {
		//if visits[id] < maxVis {
			delsVis[id] = true
		}
	}
	for k,_ := range delsVis {
		delete(cands, k)
	}
	fmt.Printf("after cands:%v\n", cands)
	for id, _ := range cands {
		res := make([]int, len(jobs))
		vis := make([]bool, len(jobs))

		depthExRec2(jobs, id, res, vis, visits, 0, &bestScore, &count, "")
	}
	log.Printf("Finished after checking %v combinations.", count)
}

func depthExRec2(jobs []types.JobNode, nodeId int, result []int, visited []bool, visits []float64, step int, bestScore *score, count *int, decisions string) {
	node := jobs[nodeId]
	result[step] = node.Id
	visited[node.Id] = true

	if step == len(jobs) - 1 {
		updateBestScoreMsg(jobs, result, step, bestScore, count, decisions)
//		os.Exit(1)
		return
	}

	cands := getCands(jobs, visited, node)

//	nextBest := -1
	nextBestScore := math.MaxFloat64
	candScores := make(map[int]float64, len(cands))

//	if minVis <= 1 {
		for id,_ := range cands {
			result[step+1] = id
			sc := evalResultStep(jobs, result, false, step+1)
			candScores[id] = sc

			if sc < nextBestScore {
//				nextBest = id
				nextBestScore = sc
			}
		}
//	}

	decisions += fmt.Sprintf("before cands:")
	for id,_ := range cands {
		decisions += fmt.Sprintf(" %v(d%v, v%v, s%v)", id, jobs[id].Depth, visits[id], candScores[id])
	}
	decisions += fmt.Sprintf("\n")

	maxDepth := 0
	for id,_ := range cands {
		if jobs[id].Depth > maxDepth {
			maxDepth = jobs[id].Depth
		}
	}
	//fmt.Printf("cands:%v md: %v\n", cands, maxDepth)
	delsDepth := make(map[int]bool)
	for id,_ := range cands {
		if jobs[id].Depth < maxDepth {
			delsDepth[id] = true
		}
	}
	for k,_ := range delsDepth {
		delete(cands, k)
	}

	minVis := math.MaxFloat32
	maxVis := 0.0
	for id, _ := range cands {
		if minVis > visits[id] {
			minVis = visits[id]
		}
		if maxVis < visits[id] {
			maxVis = visits[id]
		}
	}
	delsVis := make(map[int]bool)
	for id,_ := range cands {
//		if visits[id] > minVis {
		if visits[id] < maxVis {
			delsVis[id] = true
		}
	}
	for k,_ := range delsVis {
		delete(cands, k)
	}

	decisions += fmt.Sprintf("after cands:")
	for id,_ := range cands {
		decisions += fmt.Sprintf(" %v(d%v, v%v, s%v)", id, jobs[id].Depth, visits[id], candScores[id])
	}
	decisions += fmt.Sprintf("\n\n")

	delBest := make(map[int]bool)
	for id,_ := range cands {
		if candScores[id] > nextBestScore {
			delBest[id] = true
		}
	}
	for k,_ := range delBest {
		delete(cands, k)
	}
	for id, _ := range cands {
		resClone := make([]int, len(result))
		copy(resClone, result)
		visClone := make([]bool, len(visited))
		copy(visClone, visited)
		depthExRec2(jobs,  id, resClone, visClone, visits, step+1, bestScore, count, decisions)
	}
}

func Visits1(jobs []types.JobNode) {
	//find id of "root"
	bestScore := initScore()
	count := 0
	cands := make(map[int]bool, len(jobs))

	visits := make([]int, len(jobs))
	for _, job := range jobs {
		visitRek(job.Id, jobs, &visits)
	}
	log.Printf("visits: %v", visits)

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}

	minVis := math.MaxInt32
	for id, _ := range cands {
		if minVis > visits[id] {
			minVis = visits[id]
//			idMinVis = id
		}
	}

	for id, _ := range cands {
		if visits[id] == minVis {
			res := make([]int, len(jobs))
			vis := make([]bool, len(jobs))

			visits1Rec(jobs, id, res, vis, 0, &bestScore, &count, visits)
		}
	}
	log.Printf("Finished after checking %v combinations.", count)
}

func visits1Rec(jobs []types.JobNode, nodeId int, result []int, visited []bool, step int, bestScore *score, count *int, visits []int) {
	node := jobs[nodeId]


	result[step] = node.Id
	visited[node.Id] = true

	if step == len(jobs)-1 {
		updateBestScore(jobs, result, step, bestScore, count)
		
		return
	}

	cands := make(map[int]bool, len(jobs))

	for _, id := range node.OutIds {
		allInVis := true
		for _,inId := range jobs[id].InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[id] {
			cands[id] = true
		}
	}
	for _, job := range jobs {
		allInVis := true
		for _,inId := range job.InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[job.Id] {
			cands[job.Id] = true
		}
	}
	//log.Printf("cands: %v", cands)
//	idMinVis := -1
	minVis := math.MaxInt32
	for id, _ := range cands {
		if minVis > visits[id] {
			minVis = visits[id]
//			idMinVis = id
		}
	}

	for id, _ := range cands {
		if visits[id] == minVis {
			resClone := make([]int, len(result))
			copy(resClone, result)
			visClone := make([]bool, len(visited))
			copy(visClone, visited)
			visits1Rec(jobs, id, resClone, visClone, step+1, bestScore, count, visits)
		}
	}
//	log.Printf("step: %v res:  %v\n", step, result)
}

func Visits2(jobs []types.JobNode) {
	//find id of "root"
	bestScore := initScore()
	count := 0
	cands := make(map[int]bool, len(jobs))

	visits := make([]int, len(jobs))
	for _, job := range jobs {
		visitRek(job.Id, jobs, &visits)
	}
	log.Printf("visits: %v", visits)

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}
	minVis := math.MaxInt32
	for id, _ := range cands {
		if minVis > visits[id] {
			minVis = visits[id]
//			idMinVis = id
		}
	}

	for id, _ := range cands {
		if visits[id] == minVis {
			res := make([]int, len(jobs))
			vis := make([]bool, len(jobs))

			visits2Rec(jobs, id, res, vis, 0, &bestScore, &count, visits)
		}
	}
	log.Printf("Finished after checking %v combinations.", count)
}

func visits2Rec(jobs []types.JobNode, nodeId int, result []int, visited []bool, step int, bestScore *score, count *int, visits []int) {
	node := jobs[nodeId]


	result[step] = node.Id
	visited[node.Id] = true

	if step == len(jobs)-1 {
		updateBestScore(jobs, result, step, bestScore, count)
		
		return
	}

	cands := make(map[int]bool, len(jobs))

	for _, id := range node.OutIds {
		allInVis := true
		for _,inId := range jobs[id].InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[id] {
			cands[id] = true
		}
	}
	for _, job := range jobs {
		allInVis := true
		for _,inId := range job.InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[job.Id] {
			cands[job.Id] = true
		}
	}
	//log.Printf("cands: %v", cands)
	idMinVis := -1
	minVis := math.MaxInt32
	for id, _ := range cands {
		if minVis > visits[id] {
			minVis = visits[id]
			idMinVis = id
		}
	}

	for id, _ := range cands {
		if minVis <= 1 {
			resClone := make([]int, len(result))
			copy(resClone, result)
			visClone := make([]bool, len(visited))
			copy(visClone, visited)
			visits2Rec(jobs, idMinVis, resClone, visClone, step+1, bestScore, count, visits)
			break
		}
		if visits[id] == minVis {
			resClone := make([]int, len(result))
			copy(resClone, result)
			visClone := make([]bool, len(visited))
			copy(visClone, visited)
			visits2Rec(jobs, id, resClone, visClone, step+1, bestScore, count, visits)
		}
	}
//	log.Printf("step: %v res:  %v\n", step, result)
}

func Visits3(jobs []types.JobNode) {
	//find id of "root"
	bestScore := initScore()
	count := 0
	cands := make(map[int]bool, len(jobs))

	visits := make([]int, len(jobs))
	for _, job := range jobs {
		visitRek(job.Id, jobs, &visits)
	}
	log.Printf("visits: %v", visits)

	for _, job := range jobs {
		if len(job.Inputs) == 0 {
			cands[job.Id] = true
		}
	}
	for id, _ := range cands {
		res := make([]int, len(jobs))
		vis := make([]bool, len(jobs))

		visits3Rec(jobs, id, res, vis, 0, &bestScore, &count, visits)
	}
	log.Printf("Finished after checking %v combinations.", count)
}

func visits3Rec(jobs []types.JobNode, nodeId int, result []int, visited []bool, step int, bestScore *score, count *int, visits []int) {
	node := jobs[nodeId]


	result[step] = node.Id
	visited[node.Id] = true

	if step == len(jobs)-1 {
		updateBestScore(jobs, result, step, bestScore, count)
		/*
		*count = *count+1
		if *count%100000 == 0 {
			log.Printf("%v combinations tested. Best score: %v.", *count, bestScore.score)
		}
		//log.Printf("fin step: %v res:  %v vis:%v \n", step, result, visited)
		//fmt.Printf("fin step: %v res %v\n", step, result)
		scoreVal := evalResult(jobs, result, false)
//		fmt.Printf("'%v', %v\n", result, scoreVal)
		if scoreVal < bestScore.score {
			newBest := score{scoreVal, result}
			fmt.Printf("new best: '%v', %v\n", result, scoreVal)
			plotResult(jobs, result, true)
			_ = evalResult(jobs, result, true)
			*bestScore = newBest
		}
		*/
		return
	} else {
		//log.Printf("step: %v res:  %v vis:%v \n", step, result, visited)
	}

	cands := make(map[int]bool, len(jobs))

	for _, id := range node.OutIds {
		allInVis := true
		for _,inId := range jobs[id].InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[id] {
			cands[id] = true
		}
	}
	for _, job := range jobs {
		allInVis := true
		for _,inId := range job.InIds {
			allInVis = allInVis && visited[inId]
		}
		if allInVis && !visited[job.Id] {
			cands[job.Id] = true
		}
	}
	//log.Printf("cands: %v", cands)
	minVis := math.MaxInt32
	for id, _ := range cands {
		if minVis > visits[id] {
			minVis = visits[id]
		}
	}

	nextBest := -1
	nextBestScore := math.MaxFloat64
	candScores := make(map[int]float64, len(cands))

	for id, _ := range cands {
		result[step+1] = id
		sc := evalResultStep(jobs, result, false, step+1)
		candScores[id] = sc

		if sc < nextBestScore && visits[id] == 1 {
			nextBest = id
			nextBestScore = sc
		}
	}
	//log.Printf("minVis: %v", minVis)

	if minVis <= 1 {
		resClone := make([]int, len(result))
		copy(resClone, result)
		visClone := make([]bool, len(visited))
		copy(visClone, visited)
		visits3Rec(jobs, nextBest, resClone, visClone, step+1, bestScore, count, visits)
	} else {
		for id, _ := range cands {
			if visits[id] == minVis {
				resClone := make([]int, len(result))
				copy(resClone, result)
				visClone := make([]bool, len(visited))
				copy(visClone, visited)
				visits3Rec(jobs, id, resClone, visClone, step+1, bestScore, count, visits)
			}
		}
	}
//	log.Printf("step: %v res:  %v\n", step, result)
}
