# goremovelines
Remove empty lines in go code

## Installation
```
go get -u github.com/Eun/goremovelines/cmd/goremovelines
```

## Usage
```
usage: goremovelines [<flags>] [<path>...]

Remove empty (start / end) lines in go code

Flags:
  -h, --help             Show context-sensitive help (also try --help-long and --help-man).
  -r, --remove=func|struct|if|switch|case|for|interface|block ...  
                         Remove empty lines for the context (specify it multiple times, e.g.: -r=func -r=struct)
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

