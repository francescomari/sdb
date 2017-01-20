package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sdb [command]",
		Short: "SDB is collection of utilities for Apache Jackrabbit Oak's Segment Store",
	}
	cmd.AddCommand(newTarsCommand())
	cmd.AddCommand(newEntriesCommand())
	cmd.AddCommand(newSegmentsCommand())
	cmd.AddCommand(newSegmentCommand())
	cmd.AddCommand(newIndexCommand())
	cmd.AddCommand(newGraphCommand())
	cmd.AddCommand(newBinariesCommand())
	return cmd
}

func newTarsCommand() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "tars [dir]",
		Short: "Prints the TAR files at the provided path.",
		Run: func(cmd *cobra.Command, args []string) {
			directory, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to determine the working directory: %v.\n", err)
				os.Exit(1)
			}
			if len(args) > 1 {
				fmt.Fprintf(os.Stderr, "Too many arguments.\n")
				os.Exit(1)
			}
			if len(args) == 1 {
				directory = args[0]
			}
			if err := forEachTarFile(directory, all, doPrintTo(os.Stdout)); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to print TAR files: %v.\n", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "List both active and non-active TAR files")
	return cmd
}

func newEntriesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "entries file",
		Short: "Prints the entries from the specified TAR file.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				fmt.Fprintf(os.Stderr, "Too many arguments.\n")
				os.Exit(1)
			}
			if len(args) < 1 {
				fmt.Fprintf(os.Stderr, "Too few arguments.\n")
				os.Exit(1)
			}
			if err := forEachEntry(args[0], doPrintNameTo(os.Stdout)); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to print TAR entries: %v.\n", err)
				os.Exit(1)
			}
		},
	}
}

func newSegmentsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "segments file",
		Short: "Prints the identifiers of the segments from the specified TAR file.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				fmt.Fprintf(os.Stderr, "Too many arguments.\n")
				os.Exit(1)
			}
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Too few arguments.\n")
				os.Exit(1)
			}
			if err := forEachMatchingEntry(args[0], isAnySegment, doPrintSegmentNameTo(os.Stdout)); err != nil {
				fmt.Fprintln(os.Stderr, "Unable to print segment IDs: %v.\n", err)
				os.Exit(1)
			}
		},
	}
}

func newSegmentCommand() *cobra.Command {
	var f format
	cmd := &cobra.Command{
		Use:   "segment file id",
		Short: "Prints the identifiers of the segments from the specified TAR file.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "Too few arguments.\n")
				os.Exit(1)
			}
			if len(args) > 2 {
				fmt.Fprintf(os.Stderr, "Too many arguments.\n")
				os.Exit(1)
			}
			if err := onMatchingEntry(args[0], isSegment(args[1]), doPrintSegment(f, os.Stdout)); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to print segment: %v.\n", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().Var(&f, "format", "Output format (text, hex)")
	return cmd
}

func newIndexCommand() *cobra.Command {
	var f format
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Prints the index from the specified TAR file",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				fmt.Fprintf(os.Stderr, "Too many arguments.\n")
				os.Exit(1)
			}
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Too few arguments.\n")
				os.Exit(1)
			}
			if err := onMatchingEntry(args[0], isIndex, doPrintIndex(f, os.Stdout)); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to print the index: %v.\n", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().Var(&f, "format", "Output format (text, hex)")
	return cmd
}

func newGraphCommand() *cobra.Command {
	var f format
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Prints the graph from the specified TAR file",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				fmt.Fprintf(os.Stderr, "Too many arguments.\n")
				os.Exit(1)
			}
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Too few arguments.\n")
				os.Exit(1)
			}
			if err := onMatchingEntry(args[0], isGraph, doPrintGraph(f, os.Stdout)); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to print the graph: %v.\n", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().Var(&f, "format", "Output format (text, hex)")
	return cmd
}

func newBinariesCommand() *cobra.Command {
	var f format
	cmd := &cobra.Command{
		Use:   "binaries",
		Short: "Prints the index of binary references from the specified TAR file",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 1 {
				fmt.Fprintf(os.Stderr, "Too many arguments.\n")
				os.Exit(1)
			}
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Too few arguments.\n")
				os.Exit(1)
			}
			if err := onMatchingEntry(args[0], isBinary, doPrintBinaries(f, os.Stdout)); err != nil {
				fmt.Fprintln(os.Stderr, "Unable to print the index of binary references: %v.\n", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().Var(&f, "format", "Output format (text, hex)")
	return cmd
}
