// +build ignore

package main

import (
	"os"

	"github.com/utilizedsun/layout"
	"github.com/utilizedsun/layout/format/svg"
)

func main() {
	graph := layout.NewDigraph()
	graph.Edge("A", "B")
	graph.Edge("A", "C")
	graph.Edge("B", "D")
	graph.Edge("C", "D")

	layout.Hierarchical(graph)

	svg.Write(os.Stdout, graph)
}
