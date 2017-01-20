package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	programName = "sdb"
)

func main() {
	var t tool
	if len(os.Args) < 2 {
		commands("No command specified")
		os.Exit(1)
	}
	cmd := os.Args[1]
	var fn func([]string) error
	switch cmd {
	case "tars":
		fn = t.tars
	case "entries":
		fn = t.entries
	case "segments":
		fn = t.segments
	case "segment":
		fn = t.segment
	case "index":
		fn = t.index
	case "graph":
		fn = t.graph
	case "binaries":
		fn = t.binaries
	}
	if fn == nil {
		commands(fmt.Sprintf("Invalid command '%s'", cmd))
		os.Exit(1)
	}
	if err := fn(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func commands(reason string) {
	fmt.Fprintf(os.Stderr, "%s. Available commands:\n", reason)
	fmt.Fprintf(os.Stderr, "    tars        List active and inactive TAR files\n")
	fmt.Fprintf(os.Stderr, "    entries     List the entries of a TAR file\n")
	fmt.Fprintf(os.Stderr, "    segments    List the IDs of the segments in a TAR file\n")
	fmt.Fprintf(os.Stderr, "    segment     Print the content of a segment\n")
	fmt.Fprintf(os.Stderr, "    index       Print the content of a TAR index\n")
	fmt.Fprintf(os.Stderr, "    graph       Print the content of a TAR graph\n")
	fmt.Fprintf(os.Stderr, "    binaries    Print the content of a TAR binary index\n")
}

type tool struct {
	flags *flag.FlagSet
}

func (t *tool) tars(args []string) error {
	t.initFlags("tars", "[-all] [directory]")
	all := t.boolFlag("all", false, "List active and non-active TAR files")
	t.parseFlags(args)
	directory, err := os.Getwd()
	if err != nil {
		return err
	}
	if t.nArgs() > 0 {
		directory = t.arg(0)
	}
	return printTars(directory, *all, os.Stdout)
}

func (t *tool) entries(args []string) error {
	t.initFlags("entries", "file")
	t.parseFlags(args)
	if t.nArgs() != 1 {
		fmt.Fprintln(os.Stderr, "Invalid number of arguments")
		return nil
	}
	return printEntries(t.arg(0), os.Stdout)
}

func (t *tool) segments(args []string) error {
	t.initFlags("segments", "file")
	t.parseFlags(args)
	if t.nArgs() != 1 {
		fmt.Fprintln(os.Stderr, "Invalid number of arguments")
		return nil
	}
	return printSegments(t.arg(0), os.Stdout)
}

func (t *tool) segment(args []string) error {
	t.initFlags("segment", "[-format] file segment")
	f := t.formatFlag("format", "Output format (text, hex)")
	t.parseFlags(args)
	if t.nArgs() != 2 {
		fmt.Fprintln(os.Stderr, "Invalid number of arguments")
		return nil
	}
	return printSegment(t.arg(0), t.arg(1), *f, os.Stdout)
}

func (t *tool) index(args []string) error {
	t.initFlags("index", "[-format] file")
	f := t.formatFlag("format", "Output format (text, hex)")
	t.parseFlags(args)
	if t.nArgs() != 1 {
		fmt.Fprintln(os.Stderr, "Invalid number of arguments")
		return nil
	}
	return printIndex(t.arg(0), *f, os.Stdout)
}

func (t *tool) graph(args []string) error {
	t.initFlags("graph", "[-format] file")
	f := t.formatFlag("format", "Output format (text, hex)")
	t.parseFlags(args)
	if t.nArgs() != 1 {
		fmt.Fprintln(os.Stderr, "Invalid number of arguments")
		return nil
	}
	return printGraph(t.arg(0), *f, os.Stdout)
}

func (t *tool) binaries(args []string) error {
	t.initFlags("binaries", "[-format] file")
	f := t.formatFlag("format", "Output format (text, hex)")
	t.parseFlags(args)
	if t.nArgs() != 1 {
		fmt.Fprintln(os.Stderr, "Invalid number of arguments")
		return nil
	}
	return printBinaries(t.arg(0), *f, os.Stdout)
}

func (t *tool) initFlags(cmd, usage string) {
	t.flags = flag.NewFlagSet(cmd, flag.ContinueOnError)
	t.flags.SetOutput(os.Stderr)
	t.flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s [-help] %s\n", programName, cmd, usage)
		t.flags.PrintDefaults()
	}
}

func (t *tool) boolFlag(name string, value bool, usage string) *bool {
	return t.flags.Bool(name, value, usage)
}

func (t *tool) formatFlag(name, usage string) *format {
	f := new(format)
	t.flags.Var(f, name, usage)
	return f
}

func (t *tool) parseFlags(args []string) {
	if err := t.flags.Parse(args); err != nil {
		os.Exit(1)
	}
}

func (t *tool) nArgs() int {
	return t.flags.NArg()
}

func (t *tool) arg(i int) string {
	return t.flags.Arg(i)
}

func printTars(d string, all bool, w io.Writer) error {
	return forEachTarFile(d, all, func(n string) {
		fmt.Fprintln(w, n)
	})
}

func printBinaries(p string, f format, w io.Writer) error {
	return onMatchingEntry(p, isBinary, doPrintBinaries(f, w))
}

func printGraph(p string, f format, w io.Writer) error {
	return onMatchingEntry(p, isGraph, doPrintGraph(f, w))
}

func printIndex(p string, f format, w io.Writer) error {
	return onMatchingEntry(p, isIndex, doPrintIndex(f, w))
}

func printSegments(p string, w io.Writer) error {
	return forEachMatchingEntry(p, isAnySegment, doPrintSegmentNameTo(w))
}

func printSegment(p string, id string, f format, w io.Writer) error {
	return onMatchingEntry(p, isSegment(id), doPrintSegment(f, w))
}

func printEntries(p string, w io.Writer) error {
	return forEachEntry(p, doPrintNameTo(w))
}
