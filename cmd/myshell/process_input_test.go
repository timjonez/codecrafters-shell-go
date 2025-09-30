package main

import (
	"slices"
	"testing"
)

func TestSingleQuotes(t *testing.T) {
	input := "cat '/tmp/baz/f   69' '/tmp/baz/f   23' '/tmp/baz/f   80'"
	expected := []string{"cat", "/tmp/baz/f   69", "/tmp/baz/f   23", "/tmp/baz/f   80"}
	output := processInput(input)
	eq := slices.Compare(output, expected)
	if eq != 0 {
		t.Fatalf(`processInput returned %q; expected %q`, output, expected)
	}
}

func TestDoubleQuotes(t *testing.T) {
	input := `echo "test"  "shell's"  "world"`
	expected := []string{"echo", "test", "shell's", "world"}
	output := processInput(input)
	eq := slices.Compare(output, expected)
	if eq != 0 {
		t.Fatalf(`processInput returned %q; expected %q`, output, expected)
	}
}

func TestEscapedWhitespace(t *testing.T) {
	input := `world\ \ \ \ \ \ script`
	expected := []string{"world      script"}
	output := processInput(input)
	eq := slices.Compare(output, expected)
	if eq != 0 {
		t.Fatalf(`processInput returned %q; expected %q`, output, expected)
	}
}

func TestEscapedQuotes(t *testing.T) {
	input := `echo \'\"test script\"\'`
	expected := []string{"echo", `'"test script"'`}
	output := processInput(input)
	eq := slices.Compare(output, expected)
	if eq != 0 {
		t.Fatalf(`processInput returned %q; expected %q`, output, expected)
	}
}

func TestEscapedInDouble(t *testing.T) {
	input := `cat "file\\name" "file\ name"`
	expected := []string{"cat", `file\name`, `file\ name`}
	output := processInput(input)
	eq := slices.Compare(output, expected)
	if eq != 0 {
		t.Fatalf(`processInput returned %q; expected %q`, output, expected)
	}
}
