package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Expectation struct {
	input  []string
	output string
}

func TestPermittedCommands(t *testing.T) {
	expected :=
		`An operation to be performed against the theme.
  Valid commands are:
    upload: Add file(s) to theme [default]
    download: Download file(s) from theme
    remove: Remove file(s) from theme
    replace: Overwrite theme file(s)`

	actual := CommandDescription("upload")
	assert.Equal(t, expected, actual)
}

func TestSetupAndParseArgs(t *testing.T) {
	expectations := []Expectation{
		Expectation{input: []string{"remove", "file1", "file2", "file3"}, output: "remove"},
		Expectation{input: []string{"upload", "file1", "file2"}, output: "upload"},
		Expectation{input: []string{"file1", "file2"}, output: "download"},
		Expectation{input: []string{"--command=upload", "file1"}, output: "upload"},
		Expectation{input: []string{}, output: "download"},
	}
	for _, expectation := range expectations {
		SetupAndParseArgs(expectation.input)
		assert.Equal(t, expectation.output, command, fmt.Sprintf("%s", expectation.input))
	}
}
