# goremovelines
[![Actions Status](https://github.com/Eun/goremovelines/workflows/push/badge.svg)](https://github.com/Eun/goremovelines/actions)
[![Coverage Status](https://coveralls.io/repos/github/Eun/goremovelines/badge.svg?branch=master)](https://coveralls.io/github/Eun/goremovelines?branch=master)
[![PkgGoDev](https://img.shields.io/badge/pkg.go.dev-reference-blue)](https://pkg.go.dev/github.com/Eun/goremovelines)
[![go-report](https://goreportcard.com/badge/github.com/Eun/goremovelines)](https://goreportcard.com/report/github.com/Eun/goremovelines)
---
Remove leading / trailing blank lines in Go functions, structs, if, switches, blocks.

## Installation
```
go install github.com/Eun/goremovelines/cmd/goremovelines@latest
```

## Usage
```
usage: goremovelines [<flags>] [<path>...]

Remove leading / trailing blank lines in Go functions, structs, if, switches, blocks.

Flags:
  -h, --help             Show context-sensitive help (also try --help-long and --help-man).
  -r, --remove=func|struct|if|switch|case|for|interface|block ...  
                         Remove blank lines for the context (specify it multiple times, e.g.: --remove=func --remove=struct)
  -w, --toSource         Write result to (source) file instead of stdout
  -s, --skip=DIR... ...  Skip directories with this name when expanding '...'.
      --vendor           Enable vendoring support (skips 'vendor' directories and sets GO15VENDOREXPERIMENT=1).
  -d, --debug            Display debug messages.
  -v, --version          Show application version.

Args:
  [<path>]  Directories to format. Defaults to ".". <path>/... will recurse.
```

> It is possible to combine it with gofmt/goimport/goreturns using [gomultifmt](https://github.com/Eun/gomultifmt)

```go
package main

import "fmt"

func main() {

  fmt.Print("Hello")
  
  fmt.Print("World")

}
```
**will be transformed to**

```go
package main

import "fmt"

func main() {
	fmt.Print("Hello")

	fmt.Print("World")
}
```

## Build History
[![Build history](https://buildstats.info/github/chart/Eun/goremovelines?branch=master)](https://github.com/Eun/goremovelines/actions)