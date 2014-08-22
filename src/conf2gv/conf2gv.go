package main

import (
	"encoding/xml"
	"flag"
	"log"
	"os"
	"strings"
	"ftw.at/dbstream/server/utils"
)

type JobDefinition struct {
	Description string `xml:"description,attr"`
	Inputs      string `xml:"inputs,attr"`
	Output      string `xml:"output,attr"`
	Priority    int    `xml:"priority,attr"`
	StartTime   int64  `xml:"startTime,attr"`
	Schema      string `xml:"schema,attr"`
	Index		string `xml:"index,attr"`
	Query       string `xml:"query"`
}

type extTable struct {
	XMLName xml.Name `xml:"table"`
	Name string `xml:"name,attr"`
}

type jobNode struct {
	inputs []string
	name string
}

type Config struct {
	XMLName    xml.Name `xml:"config"`
	ExtImports []extTable `xml:"modules>module>config>tables>table"`
	Jobs       []JobDefinition `xml:"modules>module>config>jobs>job"`
}

var configFileName = flag.String("config", "", "The configuration file used.")
var format = flag.String("format", "gv", "The fomat to plot the graph in, e.g. gexf or gv.")
var stlye = flag.String("style", "id", "The style of the graph labels e.g. name, id or name_id only.")

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

func jobDef2node(jd JobDefinition) (jout jobNode) {
	inputs := csvToSlice(jd.Inputs)
	jout.inputs = make([]string, len(inputs))
	for i, inp := range inputs {
		jout.inputs[i] = utils.IOWindowFromString(inp).Name
	}

	jout.name = utils.IOWindowFromString(jd.Output).Name
	return jout
}


func main() {
	flag.Parse()
	log.SetFlags(19)
	cfg := readConfig()

	totalLen := len(cfg.Jobs) + len(cfg.ExtImports)

	jobs := make([]jobNode, totalLen)

	//ExtImports to jobNodes
	i := 0
	for _, tbl := range cfg.ExtImports {
		 jobs[i] = jobNode{name: tbl.Name}
		 i++
	}
	//jobs to jobNodes
	for _, job := range cfg.Jobs {
		 jobs[i] = jobDef2node(job)
		 i++
	}


	var writer graphWriter
	if *format == "gexf" {
		writer = gexfWriter{}
	} else if *format == "gv" {
		writer = gvWriter{}
	} else {
		log.Fatalf("format: \"%v\" not supported!", *format)
	}
	writer.writeGraph(jobs, *stlye, os.Stdout)


//	for _,job := range jobs {
//		log.Printf("Job: %+v", job)
//	}

	log.Printf("Processed %d jobs.", len(jobs))

	log.Println("Closing.")

}
