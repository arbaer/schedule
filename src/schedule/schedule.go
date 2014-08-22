package main

import (
	"algo"
	"encoding/xml"
	"eval"
	"flag"
	"fmt"
	"ftw.at/dbstream/server/utils"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
	"types"
)

type JobDefinition struct {
	Description string  `xml:"description,attr"`
	IOTime      float64 `xml:"ioTime,attr"`
	CPUTime     float64 `xml:"CpuTime,attr"`
	Inputs      string  `xml:"inputs,attr"`
	Output      string  `xml:"output,attr"`
	Priority    int     `xml:"priority,attr"`
	StartTime   int64   `xml:"startTime,attr"`
	Schema      string  `xml:"schema,attr"`
	Index       string  `xml:"index,attr"`
	Query       string  `xml:"query"`
}

type extTable struct {
	XMLName xml.Name `xml:"table"`
	IOTime  float64  `xml:"ioTime,attr"`
	Name    string   `xml:"name,attr"`
}

type Config struct {
	XMLName    xml.Name        `xml:"config"`
	ExtImports []extTable      `xml:"modules>module>config>tables>table"`
	Jobs       []JobDefinition `xml:"modules>module>config>jobs>job"`
}

var configFileName = flag.String("config", "", "The configuration file used.")
var format = flag.String("format", "", "The fomat to plot the graph in, e.g. gexf or gv.")
var useAlgo = flag.String("algo", "", "The algo used for scheduling.")
var result = flag.String("result", "", "result for evaluation only.")
var outfile = flag.String("outfile", "", "Output file.")
var rNodes = flag.Int("rnodes", 0, "Numer of random nodes.")
var rEdges = flag.Int("redges", 0, "Number of random edges.")
var rseed = flag.Int64("rseed", -1, "Number of random edges.")
var sizeBased = flag.Bool("size", false, "Defines if the size based score function should be used.")

func readConfig() (cfg Config) {
	//decode the xml config file
	cfgFile, err := os.Open(*configFileName)
	if err != nil {
		log.Fatalf("ERROR while reading config: %v\n", err)
	}
	decode := xml.NewDecoder(cfgFile)
	err = decode.Decode(&cfg)
	if err != nil {
		log.Fatalf("ERROR while decoding config: %v\n", err)
	}
	cfgFile.Close()
	return
}

func csvToSlice(csv string) (out []string) {
	if strings.Index(csv, ",") > 0 {
		out = strings.Split(csv, ",")
		for i := 0; i < len(out); i++ {
			out[i] = strings.Trim(out[i], " \t")
		}
	} else {
		if len(csv) > 0 {
			out = make([]string, 1)
			out[0] = strings.Trim(csv, " \t")
		}
	}
	return out
}

func jobDef2node(jd JobDefinition) (jout types.JobNode) {
	inputs := csvToSlice(jd.Inputs)
	jout.Inputs = make([]string, len(inputs))
	for i, inp := range inputs {
		wnd := utils.IOWindowFromString(inp)
		jout.Inputs[i] = wnd.Name
	}

	jout.Name = utils.IOWindowFromString(jd.Output).Name

	jout.CpuTime = jd.CPUTime
	if jout.CpuTime == 0.0 {
		jout.CpuTime = 1.0
	}

	jout.IOTime = jd.IOTime
	if jout.IOTime == 0.0 {
		jout.IOTime = 1.0
	}
	return jout
}

var plotFile *os.File

func main() {
	flag.Parse()
	log.SetFlags(19)

	eval.SizeBased = *sizeBased

	var jobs []types.JobNode
	var queries []string
	if *configFileName != "" {
		cfg := readConfig()
		jobs, queries = readJobs(cfg)
	} else {
		var seed int64
		if *rseed > 0 {
			seed = *rseed
		} else {
			seed = time.Now().UnixNano()
		}
		fmt.Printf("Seed: %v\n", seed)
		rand.Seed(seed)

		rn := 0
		if rNodes != nil {
			rn = *rNodes
		}
		re := 0
		if rEdges != nil {
			re = *rEdges
		}
		jobs, queries = randomJobs(rn, re)
		fmt.Printf("Random node generation.")
	}

	//f, err := os.Create("plot2.txt")
	//if err != nil {
	//	panic(err)
	//}

	//plotFile = f
	//log.Printf("file: %v", plotFile)
	//defer plotFile.Sync()
	//defer plotFile.Close()

	log.Print(jobs)

	var ordering []int

	switch *useAlgo {
	case "baseline":
		ordering = algo.Baseline(jobs)
	//case "simple2":
	//	ordering = algo.Simple2(jobs)
	//case "simple3":
	//	ordering = algo.Simple3(jobs)
	case "exhaustive":
		algo.ExhaustiveSearch(jobs)
	case "randomBreadth":
		algo.RandomBreadthFirst(jobs)
	case "breadth":
		algo.BreadthFirst(jobs)
	case "breadthEx":
		algo.BreadthEx(jobs)
	case "depthEx2":
		algo.DepthEx2(jobs)
	case "depthEx":
		algo.DepthEx(jobs)
	case "depth":
		algo.DfWalk(jobs)
	case "depthAll":
		algo.DfWalkAll(jobs)
	case "cumSubNodes":
		algo.CumSubNodes(jobs)
	case "visits1":
		algo.Visits1(jobs)
	case "visits2":
		algo.Visits2(jobs)
	case "visits3":
		algo.Visits3(jobs)
	case "greedy":
		ordering = algo.Greedy(jobs)
	case "greedyAll":
		ordering = algo.GreedyAll(jobs)
	case "score":
		ordering = algo.Score_only(jobs, *result)
	case "order":
		ordering = algo.Score_result(jobs, *result)
		//	case "medianIter":			medianIter(jobs)
		//	case "a_star":				a_star(jobs)
		//		case "a_star2":				ordering = a_star2(jobs)
		//	case "a_star3":				ordering = algo.A_star3(jobs) //HEURISTIC
	case "heuristic":
		ordering = algo.Heuristic(jobs) //HEURISTIC with size
		//	case "a_star_fast":			ordering = algo.a_star_fast(jobs) //HEURISTIC?
		//	case "a_star_mincut":		ordering = algo.a_star_mincut(jobs)
	case "a_star":
		ordering = algo.A_star_mincut_maxbw(jobs)
		//	case "a_star_mincut_depth":	algo.a_star_mincut_depth(jobs)

	default:
		log.Fatalf("Stupido!")
	}

	writeQueries(jobs, queries, ordering)

	log.Printf("Final score: %d (SUMBW), %d (WSUMBW) ordering %v", eval.Cuts_mincut_maxbw(jobs, ordering), eval.Cuts_mincut_maxbw_size(jobs, ordering), ordering)

	edgeCnt := 0
	for _, job := range jobs {
		edgeCnt += len(job.OutIds)
	}
	log.Printf("Processed %d jobs with %v edges.", len(jobs), edgeCnt)

	log.Println("Closing.")

}

func randEdge(jobCnt int) (source, target int) {
	a := int(rand.Int31n(int32(jobCnt - 1)))
	b := int(rand.Int31n(int32(jobCnt - 1)))
	if a < b {
		return a + 1, b + 1
	} else if a == b {
		return a, b + 1
	} else {
		return b + 1, a + 1
	}
	return 0, 1
}

func randomJobs(rn, re int) (jobs []types.JobNode, queries []string) {

	jobs = make([]types.JobNode, rn)

	for i := 0; i < rn; i++ {
		job := types.JobNode{}
		job.Name = fmt.Sprintf("id:%v", rn-i-1)
		job.Id = rn - i - 1
		jobs[rn-i-1] = job
	}

	edgeLookup := make(map[[2]int]bool)
	for j := re; j > 0; j-- {
		s, t := 0, 0
		newEdge := false
		for !newEdge {
			s, t = randEdge(rn)
			_, exists := edgeLookup[[2]int{s, t}]
			if !exists {
				newEdge = true
				edgeLookup[[2]int{s, t}] = true
			}
			//			fmt.Printf("edge: %v->%v\n",s ,t)
		}

		jobs[s].OutIds = append(jobs[s].OutIds, int(t))
		jobs[t].InIds = append(jobs[t].InIds, int(s))
	}
	jobs[0].OutIds = append(jobs[0].OutIds, int(1))
	jobs[1].InIds = append(jobs[1].InIds, int(0))
	queries = make([]string, rn)
	return jobs, queries
}

func readJobs(cfg Config) ([]types.JobNode, []string) {
	totalLen := len(cfg.Jobs) + len(cfg.ExtImports)

	jobs := make([]types.JobNode, totalLen)
	queries := make([]string, totalLen)

	//ExtImports to types.JobNodes
	i := 0
	for _, tbl := range cfg.ExtImports {
		ioTime := tbl.IOTime
		if ioTime == 0.0 {
			ioTime = 1.0
		}
		jobs[i] = types.JobNode{Name: tbl.Name, IOTime: ioTime}
		queries[i] = fmt.Sprintf("-- table:%v ioTime%v", tbl.Name, ioTime)
		i++
	}
	//jobs to types.JobNodes
	for _, job := range cfg.Jobs {
		jobs[i] = jobDef2node(job)
		queries[i] = job.Query
		i++
	}

	nodeIds := make(map[string]int, len(jobs))
	for i, job := range jobs {
		nodeIds[job.Name] = i
	}

	for i, job := range jobs {
		job.Id = nodeIds[job.Name]
		//		if len(job.Inputs) > 0 {
		//			job.inpIds = make([]int, len(job.Inputs))
		//		}
		if len(job.Inputs) > 0 {
			job.InIds = make([]int, len(job.Inputs))
		}
		for j, input := range job.Inputs {
			job.InIds[j] = nodeIds[input]
			parent := jobs[nodeIds[input]]
			if parent.OutIds == nil {
				parent.OutIds = make([]int, 1)
				parent.OutIds[0] = job.Id
			} else {
				parent.OutIds = append(parent.OutIds, job.Id)
			}
			jobs[nodeIds[input]] = parent
		}
		jobs[i] = job
	}
	return jobs, queries
}

func writeQueries(jobs []types.JobNode, queries []string, ordering []int) {
	posLookup := make(map[int]int, len(ordering))
	for i, id := range ordering {
		posLookup[id] = i
	}

	if ordering != nil {
		queryStr := ""
		for pos, id := range ordering {

			queryStr += fmt.Sprintf("\n-- Inputs: ")
			for _, inId := range jobs[id].InIds {
				queryStr += fmt.Sprintf("%v ", jobs[inId].Name)
			}
			queryStr += fmt.Sprintf("\n")

			queryStr += fmt.Sprintf("%v\n", strings.TrimSpace(queries[id]))
			queryStr += fmt.Sprintf("-- Cache_drops:\n")

			for _, inId := range jobs[id].InIds {
				stillNeeded := false
				for _, outId := range jobs[inId].OutIds {
					outPos, exists := posLookup[outId]
					if !exists {
						fmt.Printf("Ordering is wrong.\n")
						return
					}

					if outPos > pos {
						stillNeeded = true
					}
				}
				if stillNeeded {
					queryStr += fmt.Sprintf("-- \"%v\" still needed\n", jobs[inId].Name)
				} else {
					queryStr += fmt.Sprintf("select drop_table_cache('%v');\n", jobs[inId].Name)
				}
			}
		}
		ioutil.WriteFile(*outfile, []byte(queryStr), 0644)

		//	for _,job := range jobs {
		//		log.Printf("Job: %+v", job)
		//	}
	}
}
