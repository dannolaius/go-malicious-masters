# layout [![GoDoc](https://godoc.org/github.com/ornatedoctrin/layout?status.svg)](https://godoc.org/github.com/ornatedoctrin/layout) [![Go Report Card](https://goreportcard.com/badge/github.com/ornatedoctrin/layout)](https://goreportcard.com/report/github.com/ornatedoctrin/layout)

## Experimental

Current version and API is in experimental stage. Property names may change.

## Installation

The graph layouting can be used as a command-line tool and as a library.

To install the command-line tool:
```
go get -u github.com/ornatedoctrin/layout/cmd/glay
```

To install the package:
```
go get -u github.com/ornatedoctrin/layout
```

## Usage

Minimal usage:

```
package main

import (
    "os"

    "github.com/ornatedoctrin/layout"
    "github.com/ornatedoctrin/layout/format/svg"
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
```

![Output](./examples/minimal.png)

See other examples in `examples` folder.

## Quality

Currently the `layout.Hierarchy` algorithm output is significantly worse than graphviz. It is recommended to use `graphviz dot`, if possible.