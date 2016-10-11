package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/francescomari/sdb"
)

const (
	programName = "sdb"
)

const (
	formatHex  = "hex"
	formatText = "text"
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

	if cmd == "tars" {
		return t.tars(args)
	}

	if cmd == "entries" {
		return t.entries(args)
	}

	if cmd == "segments" {
		return t.segments(args)
	}

	if cmd == "segment" {
		return t.segment(args)
	}

	if cmd == "index" {
		return t.index(args)
	}

	if cmd == "graph" {
		return t.graph(args)
	}

	if cmd == "binaries" {
		return t.binaries(args)
	}

	t.commands(fmt.Sprintf("Invalid command '%s'", cmd))

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
	flags := flag.NewFlagSet("tars", flag.ContinueOnError)

	flags.SetOutput(t.stderr)

	flags.Usage = func() {
		fmt.Fprintf(t.stderr, "Usage: %s tars [-all] [-help] [directory]\n", programName)
		flags.PrintDefaults()
	}

	all := flags.Bool("all", false, "List active and non-active TAR files")

	if err := flags.Parse(args); err != nil {
		return nil
	}

	directory, err := os.Getwd()

	if err != nil {
		return err
	}

	if flags.NArg() > 0 {
		directory = flags.Arg(0)
	}

	return sdb.PrintTars(directory, *all, t.stdout)
}

func (t *tool) entries(args []string) error {
	flags := flag.NewFlagSet("entries", flag.ContinueOnError)

	flags.SetOutput(t.stderr)

	flags.Usage = func() {
		fmt.Fprintf(t.stderr, "Usage: %s entries file\n", programName)
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return nil
	}

	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}

	return sdb.PrintEntries(flags.Arg(0), t.stdout)
}

func (t *tool) segments(args []string) error {
	flags := flag.NewFlagSet("segments", flag.ContinueOnError)

	flags.SetOutput(t.stderr)

	flags.Usage = func() {
		fmt.Fprintf(t.stderr, "Usage: %s segments file\n", programName)
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return nil
	}

	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}

	return sdb.PrintSegments(flags.Arg(0), t.stdout)
}

func (t *tool) segment(args []string) error {
	flags := flag.NewFlagSet("segment", flag.ContinueOnError)

	flags.SetOutput(t.stderr)

	flags.Usage = func() {
		fmt.Fprintf(t.stderr, "Usage: %s segment file id\n", programName)
		flags.PrintDefaults()
	}

	format := flags.String("format", "hex", "Output format (hex, text)")

	if err := flags.Parse(args); err != nil {
		return nil
	}

	if flags.NArg() != 2 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}

	f, err := readFormat(*format)

	if err != nil {
		fmt.Fprintln(t.stderr, err)
		return nil
	}

	return sdb.PrintSegment(flags.Arg(0), flags.Arg(1), f, t.stdout)
}

func (t *tool) index(args []string) error {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)

	flags.SetOutput(t.stderr)

	flags.Usage = func() {
		fmt.Fprintf(t.stderr, "Usage: %s index file\n", programName)
		flags.PrintDefaults()
	}

	format := flags.String("format", "hex", "Output format (hex, text)")

	if err := flags.Parse(args); err != nil {
		return nil
	}

	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}

	f, err := readFormat(*format)

	if err != nil {
		fmt.Fprintln(t.stderr, err)
		return nil
	}

	return sdb.PrintIndex(flags.Arg(0), f, t.stdout)
}

func (t *tool) graph(args []string) error {
	flags := flag.NewFlagSet("graph", flag.ContinueOnError)

	flags.SetOutput(t.stderr)

	flags.Usage = func() {
		fmt.Fprintf(t.stderr, "Usage: %s graph file\n", programName)
		flags.PrintDefaults()
	}

	format := flags.String("format", "hex", "Output format (hex, text)")

	if err := flags.Parse(args); err != nil {
		return nil
	}

	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}

	f, err := readFormat(*format)

	if err != nil {
		fmt.Fprintln(t.stderr, f)
		return nil
	}

	return sdb.PrintGraph(flags.Arg(0), f, t.stdout)
}

func (t *tool) binaries(args []string) error {
	flags := flag.NewFlagSet("binaries", flag.ContinueOnError)

	flags.SetOutput(t.stderr)

	flags.Usage = func() {
		fmt.Fprintf(t.stderr, "Usage: %s binaries file\n", programName)
		flags.PrintDefaults()
	}

	format := flags.String("format", "hex", "Output format (hex, text)")

	if err := flags.Parse(args); err != nil {
		return nil
	}

	if flags.NArg() != 1 {
		fmt.Fprintln(t.stderr, "Invalid number of arguments")
		return nil
	}

	f, err := readFormat(*format)

	if err != nil {
		fmt.Fprintln(t.stderr, err)
		return nil
	}

	return sdb.PrintBinaries(flags.Arg(0), f, t.stdout)
}

func readFormat(s string) (sdb.Format, error) {
	switch s {
	case formatHex:
		return sdb.FormatHex, nil
	case formatText:
		return sdb.FormatText, nil
	default:
		return 0, fmt.Errorf("Invalid format '%s'", s)
	}
}
