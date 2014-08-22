package main 

import (
	"encoding/xml"
	"log"
	"os"
	"fmt"
)

type graphWriter interface {
	//jobNode are the nodes to write the string is the anme of a file
	writeGraph([]jobNode, string, *os.File)
}

type gexfWriter struct {
}

type gexfMain struct {
	XMLName xml.Name `xml:"gexf"`
//	xmlns string `xml:"xmlns,attr"`
	Version string `xml:"version,attr"`
	Graph gexfGraph
}

type gexfGraph struct {
	XMLName xml.Name `xml:"graph"`
	Mode string `xml:"mode,attr"`
	Defaultedgetype string `xml:"defaultedgetype,attr"`
	Nodes []gexfNode `xml:"nodes>node"`
	Edges []gexfEdge `xml:"edges>edge"`
}

type gexfNode struct {
	Id int `xml:"id,attr"`
	Label string `xml:"label,attr"`
}

type gexfEdge struct {
	Id int `xml:"id,attr"`
	Source int `xml:"source,attr"`
	Target int `xml:"target,attr"`
}

func (g gexfWriter) writeGraph(jobs []jobNode, style string, f *os.File) {
	defer f.Close()
	root := gexfMain{/*xmlns:"http://www.gexf.net/1.2draft",*/ Version:"1.2"}
	root.Graph = gexfGraph{Mode:"static", Defaultedgetype:"directed"}

	nodeIds := make(map[string]int, len(jobs))
	edgesCnt := 0

	root.Graph.Nodes = make([]gexfNode, len(jobs))
	for i, job := range jobs {
		root.Graph.Nodes[i] = gexfNode{Id:i, Label:job.name}
		nodeIds[job.name] = i
		edgesCnt += len(job.inputs)
	}

	root.Graph.Edges = make([]gexfEdge, edgesCnt)
	edgeId := 0
	for _, job := range jobs {
		for _, input := range job.inputs {
			root.Graph.Edges[edgeId] = gexfEdge{Id:edgeId, Source:nodeIds[input], Target:nodeIds[job.name]}
			edgeId++
		}
	}

	enc := xml.NewEncoder(f)
	if err := enc.Encode(root); err != nil {
		log.Panicf("error: %v\n", err)
	}
}

type gvWriter struct {}

func (g gvWriter) writeGraph(jobs []jobNode, style string, f *os.File) {
	f.WriteString(fmt.Sprintf("digraph lala {"))
	defer f.Close()

	nodeIds := make(map[string]int, len(jobs))
	for i, job := range jobs {
		nodeIds[job.name] = i
		switch style {
			case "name":
				f.WriteString(fmt.Sprintf("\t%d [label=\"%s\"];\n", i, job.name))
			case "name_id":
				f.WriteString(fmt.Sprintf("\t%d [label=\"%s (%d)\"];\n", i, job.name, i))
			case "id":
				f.WriteString(fmt.Sprintf("\t%d [label=\"%d\"];\n", i, i))
		}
	}

	edgeId := 0
	for _, job := range jobs {
		for _, input := range job.inputs {
			f.WriteString(fmt.Sprintf("\t%d -> %d;\n", nodeIds[input], nodeIds[job.name]))
			edgeId++
		}
	}
	f.WriteString("}")
}
