// You can run specific tests only using
// `go test . -only=<testname>`
package goremovelines

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var onlyTest string

func init() {
	Debug = true
	flag.StringVar(&onlyTest, "only", "", "Only run this test")
	flag.Parse()
}

func runTest(t *testing.T, test string) {
	var expectedBuffer bytes.Buffer

	inputFile := filepath.Join("_tests", test, "input.go.txt")
	expectedFile := filepath.Join("_tests", test, "expected.go.txt")

	f, err := os.Open(expectedFile)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(&expectedBuffer, f)
	if err != nil {
		panic(err)
	}
	f.Close()

	var inputBuffer bytes.Buffer
	f, err = os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(&inputBuffer, f)
	if err != nil {
		panic(err)
	}
	f.Close()

	require.NotEqual(t, expectedBuffer.String(), inputBuffer.String(), "Test files for test `%s' are the same!", test)

	var mode Mode = AllMode

	modeFile := filepath.Join("_tests", test, "mode.txt")
	f, err = os.Open(modeFile)
	if err == nil {
		var modeBuffer bytes.Buffer
		_, err = io.Copy(&modeBuffer, f)
		if err != nil {
			panic(err)
		}
		f.Close()
		var m int
		m, err = strconv.Atoi(strings.TrimSpace(modeBuffer.String()))
		if err != nil {
			panic(err)
		}
		mode = Mode(m)
	} else {
		if !os.IsNotExist(err) {
			panic(err)
		}
	}

	var cleanedBuffer bytes.Buffer
	require.NoError(t, CleanFilePath(inputFile, &cleanedBuffer, mode), "Clean for `%s' failed!", test)
	require.Equal(t, expectedBuffer.String(), cleanedBuffer.String(), "Test `%s' failed!", test)
}

func TestAllTests(t *testing.T) {
	if len(onlyTest) > 0 {
		fmt.Printf("Running `%s'\n", onlyTest)
		runTest(t, onlyTest)
		return
	}

	d, err := os.Open("_tests")
	if err != nil {
		panic(err)
	}
	defer d.Close()
	fi, err := d.Readdir(-1)
	if err != nil {
		panic(err)
	}
	for _, fi := range fi {
		if fi.Mode().IsDir() {
			runTest(t, fi.Name())
		}
	}
}

func TestFindRealStartOfBody(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		//with a bracket in front
		{"{\n\n\tHello", 2},
		{"{\n\tHello", 2},
		{"{\tHello", -1},
		// bracket and a space char before \n
		{"{\t\n\n\tHello", 3},

		// invalid
		{"H\n\n\tHello", -1},
		{"", -1},

		// with a junk in front
		{"JUNK{\n\n\tHello", -1},

		// Double brackets
		{"{\n{\nHello", 2},
	}

	for i, test := range tests {
		require.Equal(t, test.expected, findRealStartOfBody(test.input, 0, len(test.input)-1), "Test %d failed", i)
	}
}

func TestFindRealEndOfBody(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		// with a bracket in the end
		{"Hello\n\n}", 6},
		{"Hello\n}", 5},
		{"Hello}", -1},

		//bracket and a space char before }
		{"Hello\n\t}", 5},

		// invalid
		{"Hello\nH}", -1},
		{"", -1},

		// with a junk in the back
		{"Hello\n}JUNK", -1},

		// Double brackets
		{"Hello\n}\n}", 7},
	}

	for i, test := range tests {
		require.Equal(t, test.expected, findRealEndOfBody(test.input, 0, len(test.input)-1), "Test %d failed", i)
	}
}

func TestRealBody(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"{\nHello\n}", "Hello"},
		{"{\n\nHello\n\n}", "\nHello\n"},
		{"{\t\n\nHello\n\n\t}", "\nHello\n"},
	}

	for i, test := range tests {
		realStart := findRealStartOfBody(test.input, 0, len(test.input)-1)
		realEnd := findRealEndOfBody(test.input, 0, len(test.input)-1)
		realBody := test.input[realStart:realEnd]
		require.Equal(t, test.expected, realBody, "Test %d failed", i)
	}
}
