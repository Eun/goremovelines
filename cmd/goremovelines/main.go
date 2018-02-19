package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Eun/goremovelines"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	removeLineFlag    = kingpin.CommandLine.Flag("remove", "Remove empty lines for the context (specify it multiple times, e.g.: --remove=func --remove=struct)").Short('r').PlaceHolder("func|struct|if|switch|case|for|interface|block").Default("func", "struct", "if", "switch", "case", "for", "interface", "block").Strings()
	writeToSourceFlag = kingpin.CommandLine.Flag("toSource", "Write result to (source) file instead of stdout").Short('w').Default("false").Bool()
	skipFlag          = kingpin.CommandLine.Flag("skip", "Skip directories with this name when expanding '...'.").Short('s').PlaceHolder("DIR...").Strings()
	vendorFlag        = kingpin.CommandLine.Flag("vendor", "Enable vendoring support (skips 'vendor' directories and sets GO15VENDOREXPERIMENT=1).").Bool()
	debugFlag         = kingpin.CommandLine.Flag("debug", "Display debug messages.").Short('d').Bool()
)

func printMode(mode goremovelines.Mode) {
	if debugFlag == nil || !*debugFlag {
		return
	}
	debug("Mode is %d", mode)
	if mode&goremovelines.FuncMode == goremovelines.FuncMode {
		debug("> Cleaning for Funcs")
	}
	if mode&goremovelines.StructMode == goremovelines.StructMode {
		debug("> Cleaning for Structs")
	}
	if mode&goremovelines.IfMode == goremovelines.IfMode {
		debug("> Cleaning for Ifs")
	}
	if mode&goremovelines.SwitchMode == goremovelines.SwitchMode {
		debug("> Cleaning for Switches")
	}
	if mode&goremovelines.CaseMode == goremovelines.CaseMode {
		debug("> Cleaning for Cases")
	}
	if mode&goremovelines.ForMode == goremovelines.ForMode {
		debug("> Cleaning for For Loops")
	}
	if mode&goremovelines.InterfaceMode == goremovelines.InterfaceMode {
		debug("> Cleaning for Interfaces")
	}
}

func parseMode() (mode goremovelines.Mode) {
	for _, flag := range *removeLineFlag {
		switch strings.ToLower(flag) {
		case "func":
			mode = mode | goremovelines.FuncMode
		case "struct":
			mode = mode | goremovelines.StructMode
		case "if":
			mode = mode | goremovelines.IfMode
		case "switch":
			mode = mode | goremovelines.SwitchMode
		case "case":
			mode = mode | goremovelines.CaseMode
		case "for":
			mode = mode | goremovelines.ForMode
		case "interface":
			mode = mode | goremovelines.InterfaceMode
		case "block":
			mode = mode | goremovelines.BlockMode
		}
	}

	printMode(mode)

	return mode
}

func cleanPaths(paths []string, mode goremovelines.Mode) {
	for i := 0; i < len(paths); i++ {
		out := &bytes.Buffer{}
		goremovelines.Debug = *debugFlag
		if err := goremovelines.CleanFilePath(paths[i], out, mode); err != nil {
			warning(err.Error())
			continue
		}
		if writeToSourceFlag != nil && *writeToSourceFlag {
			f, err := os.Create(paths[i])
			if err == nil {
				if _, err = f.Write(out.Bytes()); err != nil {
					warning("Unable to write file `%s': %v", paths[i], err)
				}
				if err = f.Close(); err != nil {
					warning("Unable to close file `%s': %v", paths[i], err)
				}
			} else {
				warning("Unable to create file `%s': %v", paths[i], err)
			}
		} else {
			if _, err := io.Copy(os.Stdout, out); err != nil {
				warning("Unable to write to stdout (`%s'): %v", paths[i], err)
			}
		}
	}
}

func cleanPathsFromStdin(mode goremovelines.Mode) {
	in := &bytes.Buffer{}
	_, err := io.Copy(in, os.Stdin)
	if err != nil {
		warning("Unable to copy stdin: %v", err)
		return
	}

	out := &bytes.Buffer{}
	goremovelines.Debug = *debugFlag
	if err := goremovelines.CleanFile(in.String(), out, mode); err != nil {
		warning(err.Error())
		return
	}
	if writeToSourceFlag != nil && *writeToSourceFlag {
		warning("Could not write to source if reading from stdin")
	}
	if _, err := io.Copy(os.Stdout, out); err != nil {
		warning("Unable to copy to stdout: %v", err)
	}
}

func main() {
	pathsArg := kingpin.Arg("path", "Directories to format. Defaults to \".\". <path>/... will recurse.").Strings()
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.CommandLine.Version("goremovelines 1.0")
	kingpin.CommandLine.VersionFlag.Short('v')
	kingpin.CommandLine.Help = "Remove empty lines in go code"

	kingpin.Parse()

	if removeLineFlag == nil {
		log.Panic("parameter remove is nil")
	}
	if len(*removeLineFlag) <= 0 {
		return
	}

	mode := parseMode()

	if pathsArg == nil || len(*pathsArg) <= 0 {
		cleanPathsFromStdin(mode)
		return
	}

	if skipFlag == nil {
		skipFlag = &[]string{}
	}

	if os.Getenv("GO15VENDOREXPERIMENT") == "1" || (vendorFlag != nil && *vendorFlag) {
		if err := os.Setenv("GO15VENDOREXPERIMENT", "1"); err != nil {
			warning("setenv GO15VENDOREXPERIMENT: %s", err)
		}
		*skipFlag = append(*skipFlag, "vendor")
		trueValue := true
		vendorFlag = &trueValue
	}

	cleanPaths(resolvePaths(*pathsArg, *skipFlag), mode)
}
