package goremovelines

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// Mode is a bitmask that defines which lines should be removed
type Mode int

const (
	// FuncMode should be set to remove empty lines in functions
	FuncMode = 1 << iota
	// StructMode should be set to remove empty lines in structs
	StructMode = 1 << iota
	// IfMode should be set to remove empty lines in if blocks
	IfMode = 1 << iota
	// SwitchMode should be set to remove empty lines in functions
	SwitchMode = 1 << iota
	// CaseMode should be set to remove empty lines in case blocks
	CaseMode = 1 << iota
	// ForMode should be set to remove empty lines in for blocks
	ForMode = 1 << iota
	// InterfaceMode should be set to remove empty lines in interface blocks
	InterfaceMode = 1 << iota
	// BlockMode should be set to remove empty lines in blocks
	BlockMode = 1 << iota
	// AllMode includes all modes
	AllMode = FuncMode | StructMode | IfMode | SwitchMode | CaseMode | ForMode | InterfaceMode | BlockMode
)

// Debug enables/disables debug output
var Debug = false

// CleanFilePath cleans a file with the specific mode, it writes the cleaned output to `out`
func CleanFilePath(path string, out io.Writer, mode Mode) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Errorf("Unable to open `%s'", path)
	}
	var b bytes.Buffer
	_, err = io.Copy(&b, f)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return CleanFile(b.String(), out, mode)
}

// CleanFile cleans a source code with the specific mode, it writes the cleaned output to `out`
func CleanFile(src string, out io.Writer, mode Mode) error {
	_, err := clean(&src, mode)
	if err != nil {
		return err
	}
	_, err = io.WriteString(out, src)
	return err
}

func clean(src *string, mode Mode) (bool, error) {
	if Debug {
		lines := strings.Split(*src, "\n")
		for i, line := range lines {
			lines[i] = ">" + line
		}
		log.Printf("Cleaning \n%s\n", strings.Join(lines, "\n"))
	}
cleanAgain:
	set := token.NewFileSet()
	astFile, err := parser.ParseFile(set, "", *src, parser.ParseComments)
	if err != nil {
		return false, errors.Errorf("Failed to parse `%s': %v", *src, err)
	}

	for i := 0; i < len(astFile.Decls); i++ {
		mod, err := cleanNode(src, astFile.Decls[i], mode)
		if err != nil {
			return false, err
		}
		if mod {
			goto cleanAgain
		}
	}

	return false, nil
}

func findRealStartOfBody(src string, start, end int) int {
	if start < 0 || end < 0 || end <= start || end >= len(src) || len(src[start:end]) <= 0 {
		return -1
	}
	if src[start] != '{' {
		return -1
	}

	start++

	var r rune
	var width int
	for i := start; i <= end; i += width {
		r, width = utf8.DecodeRuneInString(src[i:end])
		if r == '\n' {
			return i + width
		} else if unicode.IsSpace(r) {
		} else {
			return -1
		}
	}
	return -1
}

func findRealEndOfBody(src string, start, end int) int {
	if start < 0 || end < 0 || end <= start || end >= len(src) || len(src[start:end]) <= 0 {
		return -1
	}
	if src[end] != '}' {
		return -1
	}

	var r rune
	var width int
	for i := end; i >= start; i -= width {
		r, width = utf8.DecodeLastRuneInString(src[start:i])
		if r == '\n' {
			return i - width
		} else if unicode.IsSpace(r) {
		} else {
			return -1
		}
	}
	return -1
}

func cleanSrc(src *string, start, end token.Pos, mode Mode) (bool, error) {
	startOfBody := int(start)
	endOfBody := int(end)

	for (*src)[startOfBody] != '{' {
		startOfBody--
	}
	for (*src)[endOfBody] != '}' {
		endOfBody--
	}

	if Debug {
		lines := strings.Split((*src)[startOfBody:endOfBody], "\n")
		for i, line := range lines {
			lines[i] = ">" + line
		}
		log.Printf("CleanSRC \n%s\n", strings.Join(lines, "\n"))
	}

	realStartOfBody := findRealStartOfBody(*src, startOfBody, endOfBody)
	if realStartOfBody == -1 {
		return false, nil
	}

	var r rune
	var width int

	for i := realStartOfBody; i <= endOfBody; i += width {
		r, width = utf8.DecodeRuneInString((*src)[i:endOfBody])
		if r == '\n' {
			*src = (*src)[:realStartOfBody] + (*src)[i+width:endOfBody] + (*src)[endOfBody:]
			return true, nil
		} else if unicode.IsSpace(r) {
		} else {
			break
		}
	}

	realEndOfBody := findRealEndOfBody(*src, startOfBody, endOfBody)
	if realEndOfBody == -1 {
		return false, nil
	}

	for i := realEndOfBody; i >= startOfBody; i -= width {
		r, width = utf8.DecodeLastRuneInString((*src)[startOfBody:i])
		if r == '\n' {
			*src = (*src)[:startOfBody] + (*src)[startOfBody:i-width] + (*src)[realEndOfBody:]
			return true, nil
		} else if unicode.IsSpace(r) {
		} else {
			break
		}
	}
	return false, nil
}

func cleanCase(src *string, start, end token.Pos, mode Mode, isLastCase bool) (bool, error) {
	startOfBody := int(start)
	endOfBody := int(end)

	for (*src)[startOfBody] != '\n' {
		startOfBody--
	}
	for (*src)[endOfBody] != '\n' {
		endOfBody--
	}

	findRealStartOfBodyCase := func(src string, start, end int) int {
		if start < 0 || end < 0 || end <= start || end >= len(src) || len(src[start:end]) <= 0 {
			return -1
		}

		var r rune
		var width int
		for i := start; i <= end; i += width {
			r, width = utf8.DecodeRuneInString(src[i:end])
			if r == '\n' {
				return i + width
			} else if unicode.IsSpace(r) {
			} else {
				return -1
			}
		}
		return -1
	}

	findRealEndOfBodyCase := func(src string, start, end int) int {
		if start < 0 || end < 0 || end <= start || end >= len(src) || len(src[start:end]) <= 0 {
			return -1
		}

		var r rune
		var width int
		for i := end; i >= start; i -= width {
			r, width = utf8.DecodeLastRuneInString(src[start:i])
			if r == '\n' {
				return i - width
			} else if unicode.IsSpace(r) {
			} else {
				return -1
			}
		}
		return -1
	}

	if Debug {
		lines := strings.Split((*src)[startOfBody:endOfBody], "\n")
		for i, line := range lines {
			lines[i] = ">" + line
		}
		log.Printf("CleanCase \n%s\n", strings.Join(lines, "\n"))
	}

	realStartOfBody := findRealStartOfBodyCase(*src, startOfBody, endOfBody)
	if realStartOfBody == -1 {
		return false, nil
	}

	var r rune
	var width int

	for i := realStartOfBody; i <= endOfBody; i += width {
		r, width = utf8.DecodeRuneInString((*src)[i:endOfBody])
		if r == '\n' {
			*src = (*src)[:realStartOfBody] + (*src)[i+width:endOfBody] + (*src)[endOfBody:]
			return true, nil
		} else if unicode.IsSpace(r) {
		} else {
			break
		}
	}

	var realEndOfBody int
	if isLastCase {
		realEndOfBody = findRealEndOfBodyCase(*src, startOfBody, endOfBody)
		if realEndOfBody == -1 {
			return false, nil
		}
	} else {
		realEndOfBody = endOfBody
	}

	for i := realEndOfBody; i >= startOfBody; i -= width {
		r, width = utf8.DecodeLastRuneInString((*src)[startOfBody:i])
		if r == '\n' {
			*src = (*src)[:startOfBody] + (*src)[startOfBody:i-width] + (*src)[realEndOfBody:]
			return true, nil
		} else if unicode.IsSpace(r) {
		} else {
			break
		}
	}
	return false, nil
}

func cleanNode(src *string, node interface{}, mode Mode) (bool, error) {
	switch v := node.(type) {
	case *ast.GenDecl:
		if v.Tok == token.TYPE {
			for i := 0; i < len(v.Specs); i++ {
				mod, err := cleanNode(src, v.Specs[i], mode)
				if err != nil {
					return false, err
				}
				if mod {
					return true, nil
				}
			}
		}
	case *ast.DeclStmt:
		return cleanNode(src, v.Decl, mode)
	case *ast.ExprStmt:
		return cleanNode(src, v.X, mode)
	case *ast.CallExpr:
		mod, err := cleanNode(src, v.Fun, mode)
		if err != nil {
			return false, err
		}
		if mod {
			return true, nil
		}
		for i := 0; i < len(v.Args); i++ {
			mod, err := cleanNode(src, v.Args[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	case *ast.AssignStmt:
		for i := 0; i < len(v.Lhs); i++ {
			mod, err := cleanNode(src, v.Lhs[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
		for i := 0; i < len(v.Rhs); i++ {
			mod, err := cleanNode(src, v.Rhs[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	case *ast.SelectorExpr:
		return cleanNode(src, v.X, mode)
	case *ast.BasicLit:
		if v.Kind == token.FUNC {
			return cleanSrc(src, v.Pos(), v.End(), mode)
		}
	case *ast.TypeSpec:
		return cleanNode(src, v.Type, mode)
	// funcs
	case *ast.FuncDecl:
		if v.Body == nil {
			return false, nil
		}

		if mode&FuncMode == FuncMode {
			mod, err := cleanSrc(src, v.Body.Lbrace, v.Body.Rbrace, mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}

		for i := 0; i < len(v.Body.List); i++ {
			mod, err := cleanNode(src, v.Body.List[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	case *ast.FuncLit:
		if mode&FuncMode == FuncMode {
			mod, err := cleanSrc(src, v.Body.Lbrace, v.Body.Rbrace, mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}

		for i := 0; i < len(v.Body.List); i++ {
			mod, err := cleanNode(src, v.Body.List[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	// structs
	case *ast.StructType:
		if mode&StructMode == StructMode {
			return cleanSrc(src, v.Fields.Opening, v.Fields.Closing, mode)
		}
	case *ast.CompositeLit:
		mod, err := cleanNode(src, v.Type, mode)
		if err != nil {
			return false, err
		}
		if mod {
			return true, nil
		}
		for i := 0; i < len(v.Elts); i++ {
			mod, err := cleanNode(src, v.Elts[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}

		// if this was a struct, clean the list also
		if mode&StructMode == StructMode {
			if _, ok := v.Type.(*ast.StructType); ok {
				return cleanSrc(src, v.Lbrace, v.Rbrace, mode)
			}
		}
	// if
	case *ast.IfStmt:
		if mode&IfMode == IfMode {
			mod, err := cleanSrc(src, v.Body.Lbrace, v.Body.Rbrace, mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}

			if elseBlock, ok := v.Else.(*ast.BlockStmt); ok {
				mod, err := cleanSrc(src, elseBlock.Lbrace, elseBlock.Rbrace, mode)
				if err != nil {
					return false, err
				}
				if mod {
					return true, nil
				}
			}
		}

		mod, err := cleanNode(src, v.Else, mode)
		if err != nil {
			return false, err
		}
		if mod {
			return true, nil
		}

		for i := 0; i < len(v.Body.List); i++ {
			mod, err := cleanNode(src, v.Body.List[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	// switch
	case *ast.SwitchStmt:
		if mode&SwitchMode == SwitchMode {
			mod, err := cleanSrc(src, v.Body.Lbrace, v.Body.Rbrace, mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
		lastIndex := len(v.Body.List) - 1
		for i := 0; i < len(v.Body.List); i++ {
			if caseClause, ok := v.Body.List[i].(*ast.CaseClause); ok {
				if mode&CaseMode == CaseMode {
					mod, err := cleanCase(src, caseClause.Colon, caseClause.End(), mode, i == lastIndex)
					if err != nil {
						return false, err
					}
					if mod {
						return true, nil
					}
				}
				for j := 0; j < len(caseClause.Body); j++ {
					mod, err := cleanNode(src, caseClause.Body[j], mode)
					if err != nil {
						return false, err
					}
					if mod {
						return true, nil
					}
				}
			}
			mod, err := cleanNode(src, v.Body.List[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	// for
	case *ast.ForStmt:
		if mode&ForMode == ForMode {
			mod, err := cleanSrc(src, v.Body.Lbrace, v.Body.Rbrace, mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
		for i := 0; i < len(v.Body.List); i++ {
			mod, err := cleanNode(src, v.Body.List[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	// for range
	case *ast.RangeStmt:
		if mode&ForMode == ForMode {
			mod, err := cleanSrc(src, v.Body.Lbrace, v.Body.Rbrace, mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
		for i := 0; i < len(v.Body.List); i++ {
			mod, err := cleanNode(src, v.Body.List[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	// interface
	case *ast.InterfaceType:
		if mode&InterfaceMode == InterfaceMode {
			return cleanSrc(src, v.Methods.Opening, v.Methods.Closing, mode)
		}
	// block
	case *ast.BlockStmt:
		if mode&BlockMode == BlockMode {
			mod, err := cleanSrc(src, v.Lbrace, v.Rbrace, mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
		for i := 0; i < len(v.List); i++ {
			mod, err := cleanNode(src, v.List[i], mode)
			if err != nil {
				return false, err
			}
			if mod {
				return true, nil
			}
		}
	default:
		if Debug {
			log.Printf("Unable to clean %T\n", v)
		}
	}
	return false, nil
}
