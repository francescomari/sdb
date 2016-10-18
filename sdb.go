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

type format int

const (
	formatHex format = iota
	formatText
)

func main() {
	t := tool{os.Stdin, os.Stdout, os.Stderr}
	if err := t.run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type tool struct {
	stdin  *os.File
	stdout *os.File
	stderr *os.File
}

func (t *tool) run(args []string) error {
	if len(args) < 2 {
		t.commands("No command specified")
		return nil
	}
	cmd, args := args[1], args[2:]
	switch cmd {
	case "tars":
		return t.tars(args)
	case "entries":
		return t.entries(args)
	case "segments":
		return t.segments(args)
	case "segment":
		return t.segment(args)
	case "index":
		return t.index(args)
	case "graph":
		return t.graph(args)
	case "binaries":
		return t.binaries(args)
	default:
		t.commands(fmt.Sprintf("Invalid command '%s'", cmd))
	}
	return nil
}

func (t *tool) commands(reason string) {
	fmt.Fprintf(t.stderr, "%s. Available commands:\n", reason)
	fmt.Fprintf(t.stderr, "    tars        List active and inactive TAR files\n")
	fmt.Fprintf(t.stderr, "    entries     List the entries of a TAR file\n")
	fmt.Fprintf(t.stderr, "    segments    List the IDs of the segments in a TAR file\n")
	fmt.Fprintf(t.stderr, "    segment     Print the content of a segment\n")
	fmt.Fprintf(t.stderr, "    index       Print the content of a TAR index\n")
	fmt.Fprintf(t.stderr, "    graph       Print the content of a TAR graph\n")
	fmt.Fprintf(t.stderr, "    binaries    Print the content of a TAR binary index\n")
}

func (t *tool) tars(args []string) error {
	flags := t.newFlagSet("tars", "[-all] [directory]")
	all := flags.Bool("all", false, "List active and non-active TAR files")
	t.parseFlags(flags, args)
	directory, err := os.Getwd()
	if err != nil {
		return err
	}
	if flags.NArg() > 0 {
		directory = flags.Arg(0)
	}
	return printTars(directory, *all, t.stdout)
}

func (t *tool) entries(args []string) error {
	flags := t.newFlagSet("entries", "file")
	t.parseFlags(flags, args)
	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}
	return printEntries(flags.Arg(0), t.stdout)
}

func (t *tool) segments(args []string) error {
	flags := t.newFlagSet("segments", "file")
	t.parseFlags(flags, args)
	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}
	return printSegments(flags.Arg(0), t.stdout)
}

func (t *tool) segment(args []string) error {
	flags := t.newFlagSet("segment", "[-format] file")
	format := flags.String("format", "hex", "Output format (hex, text)")
	t.parseFlags(flags, args)
	if flags.NArg() != 2 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}
	f, err := readFormat(*format)
	if err != nil {
		fmt.Fprintln(t.stderr, err)
		return nil
	}
	return printSegment(flags.Arg(0), flags.Arg(1), f, t.stdout)
}

func (t *tool) index(args []string) error {
	flags := t.newFlagSet("index", "[-format] file")
	format := flags.String("format", "hex", "Output format (hex, text)")
	t.parseFlags(flags, args)
	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}
	f, err := readFormat(*format)
	if err != nil {
		fmt.Fprintln(t.stderr, err)
		return nil
	}
	return printIndex(flags.Arg(0), f, t.stdout)
}

func (t *tool) graph(args []string) error {
	flags := t.newFlagSet("graph", "[-format] file")
	format := flags.String("format", "hex", "Output format (hex, text)")
	t.parseFlags(flags, args)
	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}
	f, err := readFormat(*format)
	if err != nil {
		fmt.Fprintln(t.stderr, f)
		return nil
	}
	return printGraph(flags.Arg(0), f, t.stdout)
}

func (t *tool) binaries(args []string) error {
	flags := t.newFlagSet("binaries", "[-format] file")
	format := flags.String("format", "hex", "Output format (hex, text)")
	t.parseFlags(flags, args)
	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}
	f, err := readFormat(*format)
	if err != nil {
		fmt.Fprintln(t.stderr, err)
		return nil
	}
	return printBinaries(flags.Arg(0), f, t.stdout)
}

func (t *tool) newFlagSet(cmd, usage string) *flag.FlagSet {
	flags := flag.NewFlagSet(cmd, flag.ContinueOnError)
	flags.SetOutput(t.stderr)
	flags.Usage = func() {
		fmt.Fprintf(t.stderr, "Usage: %s %s [-help] %s\n", programName, cmd, usage)
		flags.PrintDefaults()
	}
	return flags
}

func (t *tool) parseFlags(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
}

func readFormat(s string) (format, error) {
	switch s {
	case "hex":
		return formatHex, nil
	case "text":
		return formatText, nil
	default:
		return 0, fmt.Errorf("Invalid format '%s'", s)
	}
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
