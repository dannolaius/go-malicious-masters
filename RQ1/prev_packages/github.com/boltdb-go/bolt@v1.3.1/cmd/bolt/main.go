	package main
	
	import (
		"bytes"
		"encoding/binary"
		"errors"
		"flag"
		"fmt"
		"io"
		"io/ioutil"
		"math/rand"
		"os"
		"runtime"
		"runtime/pprof"
		"strconv"
		"strings"
		"time"
		"unicode"
		"unicode/utf8"
		"unsafe"
	
		"github.com/boltdb-go/bolt"
	)
	
	var (
	
		ErrUsage = errors.New("usage")
	
		ErrUnknownCommand = errors.New("unknown command")
	
		ErrPathRequired = errors.New("path required")
	
		ErrFileNotFound = errors.New("file not found")
	
		ErrInvalidValue = errors.New("invalid value")
	
		ErrCorrupt = errors.New("invalid value")
	
		ErrNonDivisibleBatchSize = errors.New("number of iterations must be divisible by the batch size")
	
		ErrPageIDRequired = errors.New("page id required")
	
		ErrPageNotFound = errors.New("page not found")
	
		ErrPageFreed = errors.New("page freed")
	)
	
	const PageHeaderSize = 16
	
	func main() {
		m := NewMain()
		if err := m.Run(os.Args[1:]...); err == ErrUsage {
			os.Exit(2)
		} else if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
	
	type Main struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	
	func NewMain() *Main {
		return &Main{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}
	}
	
	func (m *Main) Run(args ...string) error {
	
		if len(args) == 0 || strings.HasPrefix(args[0], "-") {
			fmt.Fprintln(m.Stderr, m.Usage())
			return ErrUsage
		}
	
		switch args[0] {
		case "help":
			fmt.Fprintln(m.Stderr, m.Usage())
			return ErrUsage
		case "bench":
			return newBenchCommand(m).Run(args[1:]...)
		case "check":
			return newCheckCommand(m).Run(args[1:]...)
		case "compact":
			return newCompactCommand(m).Run(args[1:]...)
		case "dump":
			return newDumpCommand(m).Run(args[1:]...)
		case "info":
			return newInfoCommand(m).Run(args[1:]...)
		case "page":
			return newPageCommand(m).Run(args[1:]...)
		case "pages":
			return newPagesCommand(m).Run(args[1:]...)
		case "stats":
			return newStatsCommand(m).Run(args[1:]...)
		default:
			return ErrUnknownCommand
		}
	}
	
	func (m *Main) Usage() string {
		return strings.TrimLeft(`
	Bolt is a tool for inspecting bolt databases.
	
	Usage:
	
		bolt command [arguments]
	
	The commands are:
	
	    bench       run synthetic benchmark against bolt
	    check       verifies integrity of bolt database
	    compact     copies a bolt database, compacting it in the process
	    info        print basic info
	    help        print this screen
	    pages       print list of pages with their types
	    stats       iterate over all pages and generate usage stats
	
	Use "bolt [command] -h" for more information about a command.
	`, "\n")
	}
	
	type CheckCommand struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	
	func newCheckCommand(m *Main) *CheckCommand {
		return &CheckCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		}
	}
	
	func (cmd *CheckCommand) Run(args ...string) error {
	
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		help := fs.Bool("h", false, "")
		if err := fs.Parse(args); err != nil {
			return err
		} else if *help {
			fmt.Fprintln(cmd.Stderr, cmd.Usage())
			return ErrUsage
		}
	
		path := fs.Arg(0)
		if path == "" {
			return ErrPathRequired
		} else if _, err := os.Stat(path); os.IsNotExist(err) {
			return ErrFileNotFound
		}
	
		db, err := bolt.Open(path, 0666, nil)
		if err != nil {
			return err
		}
		defer db.Close()
	
		return db.View(func(tx *bolt.Tx) error {
			var count int
			ch := tx.Check()
		loop:
			for {
				select {
				case err, ok := <-ch:
					if !ok {
						break loop
					}
					fmt.Fprintln(cmd.Stdout, err)
					count++
				}
			}
	
			if count > 0 {
				fmt.Fprintf(cmd.Stdout, "%d errors found\n", count)
				return ErrCorrupt
			}
	
			fmt.Fprintln(cmd.Stdout, "OK")
			return nil
		})
	}
	
	func (cmd *CheckCommand) Usage() string {
		return strings.TrimLeft(`
	usage: bolt check PATH
	
	Check opens a database at PATH and runs an exhaustive check to verify that
	all pages are accessible or are marked as freed. It also verifies that no
	pages are double referenced.
	
	Verification errors will stream out as they are found and the process will
	return after all pages have been checked.
	`, "\n")
	}
	
	type InfoCommand struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	
	func newInfoCommand(m *Main) *InfoCommand {
		return &InfoCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		}
	}
	
	func (cmd *InfoCommand) Run(args ...string) error {
	
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		help := fs.Bool("h", false, "")
		if err := fs.Parse(args); err != nil {
			return err
		} else if *help {
			fmt.Fprintln(cmd.Stderr, cmd.Usage())
			return ErrUsage
		}
	
		path := fs.Arg(0)
		if path == "" {
			return ErrPathRequired
		} else if _, err := os.Stat(path); os.IsNotExist(err) {
			return ErrFileNotFound
		}
	
		db, err := bolt.Open(path, 0666, nil)
		if err != nil {
			return err
		}
		defer db.Close()
	
		info := db.Info()
		fmt.Fprintf(cmd.Stdout, "Page Size: %d\n", info.PageSize)
	
		return nil
	}
	
	func (cmd *InfoCommand) Usage() string {
		return strings.TrimLeft(`
	usage: bolt info PATH
	
	Info prints basic information about the Bolt database at PATH.
	`, "\n")
	}
	
	type DumpCommand struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	
	func newDumpCommand(m *Main) *DumpCommand {
		return &DumpCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		}
	}
	
	func (cmd *DumpCommand) Run(args ...string) error {
	
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		help := fs.Bool("h", false, "")
		if err := fs.Parse(args); err != nil {
			return err
		} else if *help {
			fmt.Fprintln(cmd.Stderr, cmd.Usage())
			return ErrUsage
		}
	
		path := fs.Arg(0)
		if path == "" {
			return ErrPathRequired
		} else if _, err := os.Stat(path); os.IsNotExist(err) {
			return ErrFileNotFound
		}
	
		pageIDs, err := atois(fs.Args()[1:])
		if err != nil {
			return err
		} else if len(pageIDs) == 0 {
			return ErrPageIDRequired
		}
	
		pageSize, err := ReadPageSize(path)
		if err != nil {
			return err
		}
	
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
	
		for i, pageID := range pageIDs {
	
			if i > 0 {
				fmt.Fprintln(cmd.Stdout, "===============================================")
			}
	
			if err := cmd.PrintPage(cmd.Stdout, f, pageID, pageSize); err != nil {
				return err
			}
		}
	
		return nil
	}
	
	func (cmd *DumpCommand) PrintPage(w io.Writer, r io.ReaderAt, pageID int, pageSize int) error {
		const bytesPerLineN = 16
	
		buf := make([]byte, pageSize)
		addr := pageID * pageSize
		if n, err := r.ReadAt(buf, int64(addr)); err != nil {
			return err
		} else if n != pageSize {
			return io.ErrUnexpectedEOF
		}
	
		var prev []byte
		var skipped bool
		for offset := 0; offset < pageSize; offset += bytesPerLineN {
	
			line := buf[offset : offset+bytesPerLineN]
			isLastLine := (offset == (pageSize - bytesPerLineN))
	
			if bytes.Equal(line, prev) && !isLastLine {
				if !skipped {
					fmt.Fprintf(w, "%07x *\n", addr+offset)
					skipped = true
				}
			} else {
	
				fmt.Fprintf(w, "%07x %04x %04x %04x %04x %04x %04x %04x %04x\n", addr+offset,
					line[0:2], line[2:4], line[4:6], line[6:8],
					line[8:10], line[10:12], line[12:14], line[14:16],
				)
	
				skipped = false
			}
	
			prev = line
		}
		fmt.Fprint(w, "\n")
	
		return nil
	}
	
	func (cmd *DumpCommand) Usage() string {
		return strings.TrimLeft(`
	usage: bolt dump -page PAGEID PATH
	
	Dump prints a hexadecimal dump of a single page.
	`, "\n")
	}
	
	type PageCommand struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	
	func newPageCommand(m *Main) *PageCommand {
		return &PageCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		}
	}
	
	func (cmd *PageCommand) Run(args ...string) error {
	
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		help := fs.Bool("h", false, "")
		if err := fs.Parse(args); err != nil {
			return err
		} else if *help {
			fmt.Fprintln(cmd.Stderr, cmd.Usage())
			return ErrUsage
		}
	
		path := fs.Arg(0)
		if path == "" {
			return ErrPathRequired
		} else if _, err := os.Stat(path); os.IsNotExist(err) {
			return ErrFileNotFound
		}
	
		pageIDs, err := atois(fs.Args()[1:])
		if err != nil {
			return err
		} else if len(pageIDs) == 0 {
			return ErrPageIDRequired
		}
	
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
	
		for i, pageID := range pageIDs {
	
			if i > 0 {
				fmt.Fprintln(cmd.Stdout, "===============================================")
			}
	
			p, buf, err := ReadPage(path, pageID)
			if err != nil {
				return err
			}
	
			fmt.Fprintf(cmd.Stdout, "Page ID:    %d\n", p.id)
			fmt.Fprintf(cmd.Stdout, "Page Type:  %s\n", p.Type())
			fmt.Fprintf(cmd.Stdout, "Total Size: %d bytes\n", len(buf))
	
			switch p.Type() {
			case "meta":
				err = cmd.PrintMeta(cmd.Stdout, buf)
			case "leaf":
				err = cmd.PrintLeaf(cmd.Stdout, buf)
			case "branch":
				err = cmd.PrintBranch(cmd.Stdout, buf)
			case "freelist":
				err = cmd.PrintFreelist(cmd.Stdout, buf)
			}
			if err != nil {
				return err
			}
		}
	
		return nil
	}
	
	func (cmd *PageCommand) PrintMeta(w io.Writer, buf []byte) error {
		m := (*meta)(unsafe.Pointer(&buf[PageHeaderSize]))
		fmt.Fprintf(w, "Version:    %d\n", m.version)
		fmt.Fprintf(w, "Page Size:  %d bytes\n", m.pageSize)
		fmt.Fprintf(w, "Flags:      %08x\n", m.flags)
		fmt.Fprintf(w, "Root:       <pgid=%d>\n", m.root.root)
		fmt.Fprintf(w, "Freelist:   <pgid=%d>\n", m.freelist)
		fmt.Fprintf(w, "HWM:        <pgid=%d>\n", m.pgid)
		fmt.Fprintf(w, "Txn ID:     %d\n", m.txid)
		fmt.Fprintf(w, "Checksum:   %016x\n", m.checksum)
		fmt.Fprintf(w, "\n")
		return nil
	}
	
	func (cmd *PageCommand) PrintLeaf(w io.Writer, buf []byte) error {
		p := (*page)(unsafe.Pointer(&buf[0]))
	
		fmt.Fprintf(w, "Item Count: %d\n", p.count)
		fmt.Fprintf(w, "\n")
	
		for i := uint16(0); i < p.count; i++ {
			e := p.leafPageElement(i)
	
			var k string
			if isPrintable(string(e.key())) {
				k = fmt.Sprintf("%q", string(e.key()))
			} else {
				k = fmt.Sprintf("%x", string(e.key()))
			}
	
			var v string
			if (e.flags & uint32(bucketLeafFlag)) != 0 {
				b := (*bucket)(unsafe.Pointer(&e.value()[0]))
				v = fmt.Sprintf("<pgid=%d,seq=%d>", b.root, b.sequence)
			} else if isPrintable(string(e.value())) {
				v = fmt.Sprintf("%q", string(e.value()))
			} else {
				v = fmt.Sprintf("%x", string(e.value()))
			}
	
			fmt.Fprintf(w, "%s: %s\n", k, v)
		}
		fmt.Fprintf(w, "\n")
		return nil
	}
	
	func (cmd *PageCommand) PrintBranch(w io.Writer, buf []byte) error {
		p := (*page)(unsafe.Pointer(&buf[0]))
	
		fmt.Fprintf(w, "Item Count: %d\n", p.count)
		fmt.Fprintf(w, "\n")
	
		for i := uint16(0); i < p.count; i++ {
			e := p.branchPageElement(i)
	
			var k string
			if isPrintable(string(e.key())) {
				k = fmt.Sprintf("%q", string(e.key()))
			} else {
				k = fmt.Sprintf("%x", string(e.key()))
			}
	
			fmt.Fprintf(w, "%s: <pgid=%d>\n", k, e.pgid)
		}
		fmt.Fprintf(w, "\n")
		return nil
	}
	
	func (cmd *PageCommand) PrintFreelist(w io.Writer, buf []byte) error {
		p := (*page)(unsafe.Pointer(&buf[0]))
	
		fmt.Fprintf(w, "Item Count: %d\n", p.count)
		fmt.Fprintf(w, "\n")
	
		ids := (*[maxAllocSize]pgid)(unsafe.Pointer(&p.ptr))
		for i := uint16(0); i < p.count; i++ {
			fmt.Fprintf(w, "%d\n", ids[i])
		}
		fmt.Fprintf(w, "\n")
		return nil
	}
	
	func (cmd *PageCommand) PrintPage(w io.Writer, r io.ReaderAt, pageID int, pageSize int) error {
		const bytesPerLineN = 16
	
		buf := make([]byte, pageSize)
		addr := pageID * pageSize
		if n, err := r.ReadAt(buf, int64(addr)); err != nil {
			return err
		} else if n != pageSize {
			return io.ErrUnexpectedEOF
		}
	
		var prev []byte
		var skipped bool
		for offset := 0; offset < pageSize; offset += bytesPerLineN {
	
			line := buf[offset : offset+bytesPerLineN]
			isLastLine := (offset == (pageSize - bytesPerLineN))
	
			if bytes.Equal(line, prev) && !isLastLine {
				if !skipped {
					fmt.Fprintf(w, "%07x *\n", addr+offset)
					skipped = true
				}
			} else {
	
				fmt.Fprintf(w, "%07x %04x %04x %04x %04x %04x %04x %04x %04x\n", addr+offset,
					line[0:2], line[2:4], line[4:6], line[6:8],
					line[8:10], line[10:12], line[12:14], line[14:16],
				)
	
				skipped = false
			}
	
			prev = line
		}
		fmt.Fprint(w, "\n")
	
		return nil
	}
	
	func (cmd *PageCommand) Usage() string {
		return strings.TrimLeft(`
	usage: bolt page -page PATH pageid [pageid...]
	
	Page prints one or more pages in human readable format.
	`, "\n")
	}
	
	type PagesCommand struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	
	func newPagesCommand(m *Main) *PagesCommand {
		return &PagesCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		}
	}
	
	func (cmd *PagesCommand) Run(args ...string) error {
	
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		help := fs.Bool("h", false, "")
		if err := fs.Parse(args); err != nil {
			return err
		} else if *help {
			fmt.Fprintln(cmd.Stderr, cmd.Usage())
			return ErrUsage
		}
	
		path := fs.Arg(0)
		if path == "" {
			return ErrPathRequired
		} else if _, err := os.Stat(path); os.IsNotExist(err) {
			return ErrFileNotFound
		}
	
		db, err := bolt.Open(path, 0666, nil)
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
	
		fmt.Fprintln(cmd.Stdout, "ID       TYPE       ITEMS  OVRFLW")
		fmt.Fprintln(cmd.Stdout, "======== ========== ====== ======")
	
		return db.Update(func(tx *bolt.Tx) error {
			var id int
			for {
				p, err := tx.Page(id)
				if err != nil {
					return &PageError{ID: id, Err: err}
				} else if p == nil {
					break
				}
	
				var count, overflow string
				if p.Type != "free" {
					count = strconv.Itoa(p.Count)
					if p.OverflowCount > 0 {
						overflow = strconv.Itoa(p.OverflowCount)
					}
				}
	
				fmt.Fprintf(cmd.Stdout, "%-8d %-10s %-6s %-6s\n", p.ID, p.Type, count, overflow)
	
				id += 1
				if p.Type != "free" {
					id += p.OverflowCount
				}
			}
			return nil
		})
	}
	
	func (cmd *PagesCommand) Usage() string {
		return strings.TrimLeft(`
	usage: bolt pages PATH
	
	Pages prints a table of pages with their type (meta, leaf, branch, freelist).
	Leaf and branch pages will show a key count in the "items" column while the
	freelist will show the number of free pages in the "items" column.
	
	The "overflow" column shows the number of blocks that the page spills over
	into. Normally there is no overflow but large keys and values can cause
	a single page to take up multiple blocks.
	`, "\n")
	}
	
	type StatsCommand struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	
	func newStatsCommand(m *Main) *StatsCommand {
		return &StatsCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		}
	}
	
	func (cmd *StatsCommand) Run(args ...string) error {
	
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		help := fs.Bool("h", false, "")
		if err := fs.Parse(args); err != nil {
			return err
		} else if *help {
			fmt.Fprintln(cmd.Stderr, cmd.Usage())
			return ErrUsage
		}
	
		path, prefix := fs.Arg(0), fs.Arg(1)
		if path == "" {
			return ErrPathRequired
		} else if _, err := os.Stat(path); os.IsNotExist(err) {
			return ErrFileNotFound
		}
	
		db, err := bolt.Open(path, 0666, nil)
		if err != nil {
			return err
		}
		defer db.Close()
	
		return db.View(func(tx *bolt.Tx) error {
			var s bolt.BucketStats
			var count int
			if err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
				if bytes.HasPrefix(name, []byte(prefix)) {
					s.Add(b.Stats())
					count += 1
				}
				return nil
			}); err != nil {
				return err
			}
	
			fmt.Fprintf(cmd.Stdout, "Aggregate statistics for %d buckets\n\n", count)
	
			fmt.Fprintln(cmd.Stdout, "Page count statistics")
			fmt.Fprintf(cmd.Stdout, "\tNumber of logical branch pages: %d\n", s.BranchPageN)
			fmt.Fprintf(cmd.Stdout, "\tNumber of physical branch overflow pages: %d\n", s.BranchOverflowN)
			fmt.Fprintf(cmd.Stdout, "\tNumber of logical leaf pages: %d\n", s.LeafPageN)
			fmt.Fprintf(cmd.Stdout, "\tNumber of physical leaf overflow pages: %d\n", s.LeafOverflowN)
	
			fmt.Fprintln(cmd.Stdout, "Tree statistics")
			fmt.Fprintf(cmd.Stdout, "\tNumber of keys/value pairs: %d\n", s.KeyN)
			fmt.Fprintf(cmd.Stdout, "\tNumber of levels in B+tree: %d\n", s.Depth)
	
			fmt.Fprintln(cmd.Stdout, "Page size utilization")
			fmt.Fprintf(cmd.Stdout, "\tBytes allocated for physical branch pages: %d\n", s.BranchAlloc)
			var percentage int
			if s.BranchAlloc != 0 {
				percentage = int(float32(s.BranchInuse) * 100.0 / float32(s.BranchAlloc))
			}
			fmt.Fprintf(cmd.Stdout, "\tBytes actually used for branch data: %d (%d%%)\n", s.BranchInuse, percentage)
			fmt.Fprintf(cmd.Stdout, "\tBytes allocated for physical leaf pages: %d\n", s.LeafAlloc)
			percentage = 0
			if s.LeafAlloc != 0 {
				percentage = int(float32(s.LeafInuse) * 100.0 / float32(s.LeafAlloc))
			}
			fmt.Fprintf(cmd.Stdout, "\tBytes actually used for leaf data: %d (%d%%)\n", s.LeafInuse, percentage)
	
			fmt.Fprintln(cmd.Stdout, "Bucket statistics")
			fmt.Fprintf(cmd.Stdout, "\tTotal number of buckets: %d\n", s.BucketN)
			percentage = 0
			if s.BucketN != 0 {
				percentage = int(float32(s.InlineBucketN) * 100.0 / float32(s.BucketN))
			}
			fmt.Fprintf(cmd.Stdout, "\tTotal number on inlined buckets: %d (%d%%)\n", s.InlineBucketN, percentage)
			percentage = 0
			if s.LeafInuse != 0 {
				percentage = int(float32(s.InlineBucketInuse) * 100.0 / float32(s.LeafInuse))
			}
			fmt.Fprintf(cmd.Stdout, "\tBytes used for inlined buckets: %d (%d%%)\n", s.InlineBucketInuse, percentage)
	
			return nil
		})
	}
	
	func (cmd *StatsCommand) Usage() string {
		return strings.TrimLeft(`
	usage: bolt stats PATH
	
	Stats performs an extensive search of the database to track every page
	reference. It starts at the current meta page and recursively iterates
	through every accessible bucket.
	
	The following errors can be reported:
	
	    already freed
	        The page is referenced more than once in the freelist.
	
	    unreachable unfreed
	        The page is not referenced by a bucket or in the freelist.
	
	    reachable freed
	        The page is referenced by a bucket but is also in the freelist.
	
	    out of bounds
	        A page is referenced that is above the high water mark.
	
	    multiple references
	        A page is referenced by more than one other page.
	
	    invalid type
	        The page type is not "meta", "leaf", "branch", or "freelist".
	
	No errors should occur in your database. However, if for some reason you
	experience corruption, please submit a ticket to the Bolt project page:
	
	  https:
	`, "\n")
	}
	
	var benchBucketName = []byte("bench")
	
	type BenchCommand struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	
	func newBenchCommand(m *Main) *BenchCommand {
		return &BenchCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		}
	}
	
	func (cmd *BenchCommand) Run(args ...string) error {
	
		options, err := cmd.ParseFlags(args)
		if err != nil {
			return err
		}
	
		if options.Work {
			fmt.Fprintf(cmd.Stdout, "work: %s\n", options.Path)
		} else {
			defer os.Remove(options.Path)
		}
	
		db, err := bolt.Open(options.Path, 0666, nil)
		if err != nil {
			return err
		}
		db.NoSync = options.NoSync
		defer db.Close()
	
		var results BenchResults
		if err := cmd.runWrites(db, options, &results); err != nil {
			return fmt.Errorf("write: %v", err)
		}
	
		if err := cmd.runReads(db, options, &results); err != nil {
			return fmt.Errorf("bench: read: %s", err)
		}
	
		fmt.Fprintf(os.Stderr, "# Write\t%v\t(%v/op)\t(%v op/sec)\n", results.WriteDuration, results.WriteOpDuration(), results.WriteOpsPerSecond())
		fmt.Fprintf(os.Stderr, "# Read\t%v\t(%v/op)\t(%v op/sec)\n", results.ReadDuration, results.ReadOpDuration(), results.ReadOpsPerSecond())
		fmt.Fprintln(os.Stderr, "")
		return nil
	}
	
	func (cmd *BenchCommand) ParseFlags(args []string) (*BenchOptions, error) {
		var options BenchOptions
	
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		fs.StringVar(&options.ProfileMode, "profile-mode", "rw", "")
		fs.StringVar(&options.WriteMode, "write-mode", "seq", "")
		fs.StringVar(&options.ReadMode, "read-mode", "seq", "")
		fs.IntVar(&options.Iterations, "count", 1000, "")
		fs.IntVar(&options.BatchSize, "batch-size", 0, "")
		fs.IntVar(&options.KeySize, "key-size", 8, "")
		fs.IntVar(&options.ValueSize, "value-size", 32, "")
		fs.StringVar(&options.CPUProfile, "cpuprofile", "", "")
		fs.StringVar(&options.MemProfile, "memprofile", "", "")
		fs.StringVar(&options.BlockProfile, "blockprofile", "", "")
		fs.Float64Var(&options.FillPercent, "fill-percent", bolt.DefaultFillPercent, "")
		fs.BoolVar(&options.NoSync, "no-sync", false, "")
		fs.BoolVar(&options.Work, "work", false, "")
		fs.StringVar(&options.Path, "path", "", "")
		fs.SetOutput(cmd.Stderr)
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
	
		if options.BatchSize == 0 {
			options.BatchSize = options.Iterations
		} else if options.Iterations%options.BatchSize != 0 {
			return nil, ErrNonDivisibleBatchSize
		}
	
		if options.Path == "" {
			f, err := ioutil.TempFile("", "bolt-bench-")
			if err != nil {
				return nil, fmt.Errorf("temp file: %s", err)
			}
			f.Close()
			os.Remove(f.Name())
			options.Path = f.Name()
		}
	
		return &options, nil
	}
	
	func (cmd *BenchCommand) runWrites(db *bolt.DB, options *BenchOptions, results *BenchResults) error {
	
		if options.ProfileMode == "rw" || options.ProfileMode == "w" {
			cmd.startProfiling(options)
		}
	
		t := time.Now()
	
		var err error
		switch options.WriteMode {
		case "seq":
			err = cmd.runWritesSequential(db, options, results)
		case "rnd":
			err = cmd.runWritesRandom(db, options, results)
		case "seq-nest":
			err = cmd.runWritesSequentialNested(db, options, results)
		case "rnd-nest":
			err = cmd.runWritesRandomNested(db, options, results)
		default:
			return fmt.Errorf("invalid write mode: %s", options.WriteMode)
		}
	
		results.WriteDuration = time.Since(t)
	
		if options.ProfileMode == "w" {
			cmd.stopProfiling()
		}
	
		return err
	}
	
	func (cmd *BenchCommand) runWritesSequential(db *bolt.DB, options *BenchOptions, results *BenchResults) error {
		var i = uint32(0)
		return cmd.runWritesWithSource(db, options, results, func() uint32 { i++; return i })
	}
	
	func (cmd *BenchCommand) runWritesRandom(db *bolt.DB, options *BenchOptions, results *BenchResults) error {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		return cmd.runWritesWithSource(db, options, results, func() uint32 { return r.Uint32() })
	}
	
	func (cmd *BenchCommand) runWritesSequentialNested(db *bolt.DB, options *BenchOptions, results *BenchResults) error {
		var i = uint32(0)
		return cmd.runWritesWithSource(db, options, results, func() uint32 { i++; return i })
	}
	
	func (cmd *BenchCommand) runWritesRandomNested(db *bolt.DB, options *BenchOptions, results *BenchResults) error {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		return cmd.runWritesWithSource(db, options, results, func() uint32 { return r.Uint32() })
	}
	
	func (cmd *BenchCommand) runWritesWithSource(db *bolt.DB, options *BenchOptions, results *BenchResults, keySource func() uint32) error {
		results.WriteOps = options.Iterations
	
		for i := 0; i < options.Iterations; i += options.BatchSize {
			if err := db.Update(func(tx *bolt.Tx) error {
				b, _ := tx.CreateBucketIfNotExists(benchBucketName)
				b.FillPercent = options.FillPercent
	
				for j := 0; j < options.BatchSize; j++ {
					key := make([]byte, options.KeySize)
					value := make([]byte, options.ValueSize)
	
					binary.BigEndian.PutUint32(key, keySource())
	
					if err := b.Put(key, value); err != nil {
						return err
					}
				}
	
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}
	
	func (cmd *BenchCommand) runWritesNestedWithSource(db *bolt.DB, options *BenchOptions, results *BenchResults, keySource func() uint32) error {
		results.WriteOps = options.Iterations
	
		for i := 0; i < options.Iterations; i += options.BatchSize {
			if err := db.Update(func(tx *bolt.Tx) error {
				top, err := tx.CreateBucketIfNotExists(benchBucketName)
				if err != nil {
					return err
				}
				top.FillPercent = options.FillPercent
	
				name := make([]byte, options.KeySize)
				binary.BigEndian.PutUint32(name, keySource())
	
				b, err := top.CreateBucketIfNotExists(name)
				if err != nil {
					return err
				}
				b.FillPercent = options.FillPercent
	
				for j := 0; j < options.BatchSize; j++ {
					var key = make([]byte, options.KeySize)
					var value = make([]byte, options.ValueSize)
	
					binary.BigEndian.PutUint32(key, keySource())
	
					if err := b.Put(key, value); err != nil {
						return err
					}
				}
	
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}
	
	func (cmd *BenchCommand) runReads(db *bolt.DB, options *BenchOptions, results *BenchResults) error {
	
		if options.ProfileMode == "r" {
			cmd.startProfiling(options)
		}
	
		t := time.Now()
	
		var err error
		switch options.ReadMode {
		case "seq":
			switch options.WriteMode {
			case "seq-nest", "rnd-nest":
				err = cmd.runReadsSequentialNested(db, options, results)
			default:
				err = cmd.runReadsSequential(db, options, results)
			}
		default:
			return fmt.Errorf("invalid read mode: %s", options.ReadMode)
		}
	
		results.ReadDuration = time.Since(t)
	
		if options.ProfileMode == "rw" || options.ProfileMode == "r" {
			cmd.stopProfiling()
		}
	
		return err
	}
	
	func (cmd *BenchCommand) runReadsSequential(db *bolt.DB, options *BenchOptions, results *BenchResults) error {
		return db.View(func(tx *bolt.Tx) error {
			t := time.Now()
	
			for {
				var count int
	
				c := tx.Bucket(benchBucketName).Cursor()
				for k, v := c.First(); k != nil; k, v = c.Next() {
					if v == nil {
						return errors.New("invalid value")
					}
					count++
				}
	
				if options.WriteMode == "seq" && count != options.Iterations {
					return fmt.Errorf("read seq: iter mismatch: expected %d, got %d", options.Iterations, count)
				}
	
				results.ReadOps += count
	
				if time.Since(t) >= time.Second {
					break
				}
			}
	
			return nil
		})
	}
	
	func (cmd *BenchCommand) runReadsSequentialNested(db *bolt.DB, options *BenchOptions, results *BenchResults) error {
		return db.View(func(tx *bolt.Tx) error {
			t := time.Now()
	
			for {
				var count int
				var top = tx.Bucket(benchBucketName)
				if err := top.ForEach(func(name, _ []byte) error {
					c := top.Bucket(name).Cursor()
					for k, v := c.First(); k != nil; k, v = c.Next() {
						if v == nil {
							return ErrInvalidValue
						}
						count++
					}
					return nil
				}); err != nil {
					return err
				}
	
				if options.WriteMode == "seq-nest" && count != options.Iterations {
					return fmt.Errorf("read seq-nest: iter mismatch: expected %d, got %d", options.Iterations, count)
				}
	
				results.ReadOps += count
	
				if time.Since(t) >= time.Second {
					break
				}
			}
	
			return nil
		})
	}
	
	var cpuprofile, memprofile, blockprofile *os.File
	
	func (cmd *BenchCommand) startProfiling(options *BenchOptions) {
		var err error
	
		if options.CPUProfile != "" {
			cpuprofile, err = os.Create(options.CPUProfile)
			if err != nil {
				fmt.Fprintf(cmd.Stderr, "bench: could not create cpu profile %q: %v\n", options.CPUProfile, err)
				os.Exit(1)
			}
			pprof.StartCPUProfile(cpuprofile)
		}
	
		if options.MemProfile != "" {
			memprofile, err = os.Create(options.MemProfile)
			if err != nil {
				fmt.Fprintf(cmd.Stderr, "bench: could not create memory profile %q: %v\n", options.MemProfile, err)
				os.Exit(1)
			}
			runtime.MemProfileRate = 4096
		}
	
		if options.BlockProfile != "" {
			blockprofile, err = os.Create(options.BlockProfile)
			if err != nil {
				fmt.Fprintf(cmd.Stderr, "bench: could not create block profile %q: %v\n", options.BlockProfile, err)
				os.Exit(1)
			}
			runtime.SetBlockProfileRate(1)
		}
	}
	
	func (cmd *BenchCommand) stopProfiling() {
		if cpuprofile != nil {
			pprof.StopCPUProfile()
			cpuprofile.Close()
			cpuprofile = nil
		}
	
		if memprofile != nil {
			pprof.Lookup("heap").WriteTo(memprofile, 0)
			memprofile.Close()
			memprofile = nil
		}
	
		if blockprofile != nil {
			pprof.Lookup("block").WriteTo(blockprofile, 0)
			blockprofile.Close()
			blockprofile = nil
			runtime.SetBlockProfileRate(0)
		}
	}
	
	type BenchOptions struct {
		ProfileMode   string
		WriteMode     string
		ReadMode      string
		Iterations    int
		BatchSize     int
		KeySize       int
		ValueSize     int
		CPUProfile    string
		MemProfile    string
		BlockProfile  string
		StatsInterval time.Duration
		FillPercent   float64
		NoSync        bool
		Work          bool
		Path          string
	}
	
	type BenchResults struct {
		WriteOps      int
		WriteDuration time.Duration
		ReadOps       int
		ReadDuration  time.Duration
	}
	
	func (r *BenchResults) WriteOpDuration() time.Duration {
		if r.WriteOps == 0 {
			return 0
		}
		return r.WriteDuration / time.Duration(r.WriteOps)
	}
	
	func (r *BenchResults) WriteOpsPerSecond() int {
		var op = r.WriteOpDuration()
		if op == 0 {
			return 0
		}
		return int(time.Second) / int(op)
	}
	
	func (r *BenchResults) ReadOpDuration() time.Duration {
		if r.ReadOps == 0 {
			return 0
		}
		return r.ReadDuration / time.Duration(r.ReadOps)
	}
	
	func (r *BenchResults) ReadOpsPerSecond() int {
		var op = r.ReadOpDuration()
		if op == 0 {
			return 0
		}
		return int(time.Second) / int(op)
	}
	
	type PageError struct {
		ID  int
		Err error
	}
	
	func (e *PageError) Error() string {
		return fmt.Sprintf("page error: id=%d, err=%s", e.ID, e.Err)
	}
	
	func isPrintable(s string) bool {
		if !utf8.ValidString(s) {
			return false
		}
		for _, ch := range s {
			if !unicode.IsPrint(ch) {
				return false
			}
		}
		return true
	}
	
	func ReadPage(path string, pageID int) (*page, []byte, error) {
	
		pageSize, err := ReadPageSize(path)
		if err != nil {
			return nil, nil, fmt.Errorf("read page size: %s", err)
		}
	
		f, err := os.Open(path)
		if err != nil {
			return nil, nil, err
		}
		defer f.Close()
	
		buf := make([]byte, pageSize)
		if n, err := f.ReadAt(buf, int64(pageID*pageSize)); err != nil {
			return nil, nil, err
		} else if n != len(buf) {
			return nil, nil, io.ErrUnexpectedEOF
		}
	
		p := (*page)(unsafe.Pointer(&buf[0]))
		overflowN := p.overflow
	
		buf = make([]byte, (int(overflowN)+1)*pageSize)
		if n, err := f.ReadAt(buf, int64(pageID*pageSize)); err != nil {
			return nil, nil, err
		} else if n != len(buf) {
			return nil, nil, io.ErrUnexpectedEOF
		}
		p = (*page)(unsafe.Pointer(&buf[0]))
	
		return p, buf, nil
	}
	
	func ReadPageSize(path string) (int, error) {
	
		f, err := os.Open(path)
		if err != nil {
			return 0, err
		}
		defer f.Close()
	
		buf := make([]byte, 4096)
		if _, err := io.ReadFull(f, buf); err != nil {
			return 0, err
		}
	
		m := (*meta)(unsafe.Pointer(&buf[PageHeaderSize]))
		return int(m.pageSize), nil
	}
	
	func atois(strs []string) ([]int, error) {
		var a []int
		for _, str := range strs {
			i, err := strconv.Atoi(str)
			if err != nil {
				return nil, err
			}
			a = append(a, i)
		}
		return a, nil
	}
	
	const maxAllocSize = 0xFFFFFFF
	
	const (
		branchPageFlag   = 0x01
		leafPageFlag     = 0x02
		metaPageFlag     = 0x04
		freelistPageFlag = 0x10
	)
	
	const bucketLeafFlag = 0x01
	
	type pgid uint64
	
	type txid uint64
	
	type meta struct {
		magic    uint32
		version  uint32
		pageSize uint32
		flags    uint32
		root     bucket
		freelist pgid
		pgid     pgid
		txid     txid
		checksum uint64
	}
	
	type bucket struct {
		root     pgid
		sequence uint64
	}
	
	type page struct {
		id       pgid
		flags    uint16
		count    uint16
		overflow uint32
		ptr      uintptr
	}
	
	func (p *page) Type() string {
		if (p.flags & branchPageFlag) != 0 {
			return "branch"
		} else if (p.flags & leafPageFlag) != 0 {
			return "leaf"
		} else if (p.flags & metaPageFlag) != 0 {
			return "meta"
		} else if (p.flags & freelistPageFlag) != 0 {
			return "freelist"
		}
		return fmt.Sprintf("unknown<%02x>", p.flags)
	}
	
	func (p *page) leafPageElement(index uint16) *leafPageElement {
		n := &((*[0x7FFFFFF]leafPageElement)(unsafe.Pointer(&p.ptr)))[index]
		return n
	}
	
	func (p *page) branchPageElement(index uint16) *branchPageElement {
		return &((*[0x7FFFFFF]branchPageElement)(unsafe.Pointer(&p.ptr)))[index]
	}
	
	type branchPageElement struct {
		pos   uint32
		ksize uint32
		pgid  pgid
	}
	
	func (n *branchPageElement) key() []byte {
		buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
		return buf[n.pos : n.pos+n.ksize]
	}
	
	type leafPageElement struct {
		flags uint32
		pos   uint32
		ksize uint32
		vsize uint32
	}
	
	func (n *leafPageElement) key() []byte {
		buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
		return buf[n.pos : n.pos+n.ksize]
	}
	
	func (n *leafPageElement) value() []byte {
		buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
		return buf[n.pos+n.ksize : n.pos+n.ksize+n.vsize]
	}
	
	type CompactCommand struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	
		SrcPath   string
		DstPath   string
		TxMaxSize int64
	}
	
	func newCompactCommand(m *Main) *CompactCommand {
		return &CompactCommand{
			Stdin:  m.Stdin,
			Stdout: m.Stdout,
			Stderr: m.Stderr,
		}
	}
	
	func (cmd *CompactCommand) Run(args ...string) (err error) {
	
		fs := flag.NewFlagSet("", flag.ContinueOnError)
		fs.SetOutput(ioutil.Discard)
		fs.StringVar(&cmd.DstPath, "o", "", "")
		fs.Int64Var(&cmd.TxMaxSize, "tx-max-size", 65536, "")
		if err := fs.Parse(args); err == flag.ErrHelp {
			fmt.Fprintln(cmd.Stderr, cmd.Usage())
			return ErrUsage
		} else if err != nil {
			return err
		} else if cmd.DstPath == "" {
			return fmt.Errorf("output file required")
		}
	
		cmd.SrcPath = fs.Arg(0)
		if cmd.SrcPath == "" {
			return ErrPathRequired
		}
	
		fi, err := os.Stat(cmd.SrcPath)
		if os.IsNotExist(err) {
			return ErrFileNotFound
		} else if err != nil {
			return err
		}
		initialSize := fi.Size()
	
		src, err := bolt.Open(cmd.SrcPath, 0444, nil)
		if err != nil {
			return err
		}
		defer src.Close()
	
		dst, err := bolt.Open(cmd.DstPath, fi.Mode(), nil)
		if err != nil {
			return err
		}
		defer dst.Close()
	
		if err := cmd.compact(dst, src); err != nil {
			return err
		}
	
		fi, err = os.Stat(cmd.DstPath)
		if err != nil {
			return err
		} else if fi.Size() == 0 {
			return fmt.Errorf("zero db size")
		}
		fmt.Fprintf(cmd.Stdout, "%d -> %d bytes (gain=%.2fx)\n", initialSize, fi.Size(), float64(initialSize)/float64(fi.Size()))
	
		return nil
	}
	
	func (cmd *CompactCommand) compact(dst, src *bolt.DB) error {
	
		var size int64
		tx, err := dst.Begin(true)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	
		if err := cmd.walk(src, func(keys [][]byte, k, v []byte, seq uint64) error {
	
			sz := int64(len(k) + len(v))
			if size+sz > cmd.TxMaxSize && cmd.TxMaxSize != 0 {
	
				if err := tx.Commit(); err != nil {
					return err
				}
	
				tx, err = dst.Begin(true)
				if err != nil {
					return err
				}
				size = 0
			}
			size += sz
	
			nk := len(keys)
			if nk == 0 {
				bkt, err := tx.CreateBucket(k)
				if err != nil {
					return err
				}
				if err := bkt.SetSequence(seq); err != nil {
					return err
				}
				return nil
			}
	
			b := tx.Bucket(keys[0])
			if nk > 1 {
				for _, k := range keys[1:] {
					b = b.Bucket(k)
				}
			}
	
			if v == nil {
				bkt, err := b.CreateBucket(k)
				if err != nil {
					return err
				}
				if err := bkt.SetSequence(seq); err != nil {
					return err
				}
				return nil
			}
	
			return b.Put(k, v)
		}); err != nil {
			return err
		}
	
		return tx.Commit()
	}
	
	type walkFunc func(keys [][]byte, k, v []byte, seq uint64) error
	
	func (cmd *CompactCommand) walk(db *bolt.DB, walkFn walkFunc) error {
		return db.View(func(tx *bolt.Tx) error {
			return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
				return cmd.walkBucket(b, nil, name, nil, b.Sequence(), walkFn)
			})
		})
	}
	
	func (cmd *CompactCommand) walkBucket(b *bolt.Bucket, keypath [][]byte, k, v []byte, seq uint64, fn walkFunc) error {
	
		if err := fn(keypath, k, v, seq); err != nil {
			return err
		}
	
		if v != nil {
			return nil
		}
	
		keypath = append(keypath, k)
		return b.ForEach(func(k, v []byte) error {
			if v == nil {
				bkt := b.Bucket(k)
				return cmd.walkBucket(bkt, keypath, k, nil, bkt.Sequence(), fn)
			}
			return cmd.walkBucket(b, keypath, k, v, b.Sequence(), fn)
		})
	}
	
	func (cmd *CompactCommand) Usage() string {
		return strings.TrimLeft(`
	usage: bolt compact [options] -o DST SRC
	
	Compact opens a database at SRC path and walks it recursively, copying keys
	as they are found from all buckets, to a newly created database at DST path.
	
	The original database is left untouched.
	
	Additional options include:
	
		-tx-max-size NUM
			Specifies the maximum size of individual transactions.
			Defaults to 64KB.
	`, "\n")
	}
	