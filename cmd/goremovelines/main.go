package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Eun/goremovelines"
	"github.com/alecthomas/kingpin/v2"
)

// this info gets filled by goreleaser

var version string
var commit string
var date string

var (
	removeLineFlag = kingpin.CommandLine.Flag(
		"remove",
		"Remove blank lines for the context (specify it multiple times, e.g.: --remove=func --remove=struct)",
	).
		Short('r').
		PlaceHolder("func|struct|if|switch|case|for|interface|block").
		Default("func", "struct", "if", "switch", "case", "for", "interface", "block").
		Strings()
	writeToSourceFlag = kingpin.CommandLine.Flag(
		"toSource",
		"Write result to (source) file instead of stdout",
	).
		Short('w').
		Default("false").
		Bool()
	skipFlag = kingpin.CommandLine.Flag(
		"skip",
		"Skip directories with this name when expanding '...'.",
	).
		Short('s').
		PlaceHolder("DIR...").
		Strings()
	vendorFlag = kingpin.CommandLine.Flag(
		"vendor",
		"Enable vendoring support (skips 'vendor' directories and sets GO15VENDOREXPERIMENT=1).",
	).
		Bool()
	debugFlag = kingpin.CommandLine.Flag(
		"debug",
		"Display debug messages.",
	).
		Short('d').
		Bool()
)

func printMode(mode goremovelines.Mode) {
	if debugFlag == nil || !*debugFlag {
		return
	}
	debugf("Mode is %d", mode)
	if mode&goremovelines.FuncMode == goremovelines.FuncMode {
		debugf("> Cleaning for Funcs")
	}
	if mode&goremovelines.StructMode == goremovelines.StructMode {
		debugf("> Cleaning for Structs")
	}
	if mode&goremovelines.IfMode == goremovelines.IfMode {
		debugf("> Cleaning for Ifs")
	}
	if mode&goremovelines.SwitchMode == goremovelines.SwitchMode {
		debugf("> Cleaning for Switches")
	}
	if mode&goremovelines.CaseMode == goremovelines.CaseMode {
		debugf("> Cleaning for Cases")
	}
	if mode&goremovelines.ForMode == goremovelines.ForMode {
		debugf("> Cleaning for For Loops")
	}
	if mode&goremovelines.InterfaceMode == goremovelines.InterfaceMode {
		debugf("> Cleaning for Interfaces")
	}
}

func parseMode() (mode goremovelines.Mode) {
	for _, flag := range *removeLineFlag {
		switch strings.ToLower(flag) {
		case "func":
			mode |= goremovelines.FuncMode
		case "struct":
			mode |= goremovelines.StructMode
		case "if":
			mode |= goremovelines.IfMode
		case "switch":
			mode |= goremovelines.SwitchMode
		case "case":
			mode |= goremovelines.CaseMode
		case "for":
			mode |= goremovelines.ForMode
		case "interface":
			mode |= goremovelines.InterfaceMode
		case "block":
			mode |= goremovelines.BlockMode
		}
	}

	printMode(mode)

	return mode
}

func cleanPaths(paths []string, mode goremovelines.Mode) error {
	for i := 0; i < len(paths); i++ {
		out := &bytes.Buffer{}
		goremovelines.Debug = *debugFlag
		if err := goremovelines.CleanFilePath(paths[i], out, mode); err != nil {
			return err
		}
		if writeToSourceFlag != nil && *writeToSourceFlag {
			f, err := os.Create(paths[i])
			if err == nil {
				if _, err = f.Write(out.Bytes()); err != nil {
					return fmt.Errorf("unable to write file `%s': %w", paths[i], err)
				}
				if err = f.Close(); err != nil {
					return fmt.Errorf("unable to close file `%s': %w", paths[i], err)
				}
			} else {
				return fmt.Errorf("unable to create file `%s': %w", paths[i], err)
			}
		} else {
			if _, err := io.Copy(os.Stdout, out); err != nil {
				return fmt.Errorf("unable to write to stdout (`%s'): %w", paths[i], err)
			}
		}
	}
	return nil
}

func cleanPathsFromStdin(mode goremovelines.Mode) error {
	in := &bytes.Buffer{}
	_, err := io.Copy(in, os.Stdin)
	if err != nil {
		return fmt.Errorf("unable to copy stdin: %w", err)
	}

	out := &bytes.Buffer{}
	goremovelines.Debug = *debugFlag
	if err := goremovelines.CleanFile(in.String(), out, mode); err != nil {
		return err
	}
	if writeToSourceFlag != nil && *writeToSourceFlag {
		return errors.New("could not write to source if reading from stdin")
	}
	if _, err := io.Copy(os.Stdout, out); err != nil {
		return fmt.Errorf("unable to copy to stdout: %w", err)
	}
	return nil
}

func main() {
	pathsArg := kingpin.Arg("path", "Directories to format. Defaults to \".\". <path>/... will recurse.").Strings()
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.CommandLine.Version(fmt.Sprintf("goremovelines %s %s %s", version, commit, date))
	kingpin.CommandLine.VersionFlag.Short('v')
	kingpin.CommandLine.Help = "Remove leading / trailing blank lines in Go functions, structs, if, switches, blocks."

	kingpin.Parse()

	if removeLineFlag == nil {
		log.Panic("parameter remove is nil")
	}
	if len(*removeLineFlag) == 0 {
		return
	}

	mode := parseMode()

	if pathsArg == nil || len(*pathsArg) == 0 {
		if err := cleanPathsFromStdin(mode); err != nil {
			warningf("Unable to clean: %v", err.Error())
			os.Exit(1)
		}
		return
	}

	if skipFlag == nil {
		skipFlag = &[]string{}
	}

	if os.Getenv("GO15VENDOREXPERIMENT") == "1" || (vendorFlag != nil && *vendorFlag) {
		if err := os.Setenv("GO15VENDOREXPERIMENT", "1"); err != nil {
			warningf("setenv GO15VENDOREXPERIMENT: %s", err)
		}
		*skipFlag = append(*skipFlag, "vendor")
		trueValue := true
		vendorFlag = &trueValue
	}

	if err := cleanPaths(resolvePaths(*pathsArg, *skipFlag), mode); err != nil {
		warningf("Unable to clean: %v", err.Error())
		os.Exit(1)
	}
}
