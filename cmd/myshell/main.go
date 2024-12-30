package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

const SuccessCode = 0
const ErrCode = 1

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprint(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)
		commands := processInput(input)

		switch commands[0] {
		case "exit":
			if len(commands) == 1 {
				os.Exit(SuccessCode)
			}
			code, err := strconv.Atoi(commands[1])
			if err != nil {
				os.Exit(ErrCode)
			}
			os.Exit(code)
		case "echo":
			cmd := strings.Join(commands[1:], " ")
			fmt.Println(cmd)
		case "pwd":
			out, err := os.Getwd()
			if err != nil {
				fmt.Fprint(os.Stderr, "pwd: not found")
			}
			fmt.Println(out)
		case "type":
			cmd := commands[1]
			if slices.Contains([]string{"exit", "echo", "type", "pwd"}, cmd) {
				fmt.Fprint(os.Stdout, cmd+" is a shell builtin")
			} else if handleTypeCommand(commands[1:]) {
				continue
			} else {
				fmt.Fprint(os.Stderr, cmd+": not found")
			}
		case "cd":
			cmd := commands[1]
			final := cmd
			if strings.Contains(final, "~") {
				final = strings.ReplaceAll(final, "~", os.Getenv("HOME"))
			}
			if err := os.Chdir(final); err != nil {
				fmt.Fprint(os.Stderr, "cd: "+cmd+": No such file or directory")
			}
		default:
			execFile(commands)
		}
	}
}

func findFileOnPath(name string) string {
	dirs := strings.Split(os.Getenv("PATH"), ":")
	for _, dir := range dirs {
		file_path := dir + "/" + name
		_, err := os.Stat(file_path)
		if err == nil {
			return file_path
		}
	}
	return ""
}

func handleTypeCommand(args []string) bool {
	cmd := args[0]
	if file := findFileOnPath(cmd); file != "" {
		fmt.Println(cmd + " is " + file)
		return true
	}
	return false
}

func execFile(args []string) error {
	cmd := args[0]
	file := findFileOnPath(cmd)
	if file == "" {
		fmt.Fprint(os.Stderr, cmd+": command not found")
	}
	command := exec.Command(file, args[1:]...)
	output, err := command.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Fprint(os.Stdout, string(output))
	return nil
}

type InputType int

const (
	Normal InputType = iota
	SingleQuote
	DoubleQuote
)

func processInput(message string) []string {
	result := []string{}
	current := ""
	inputState := Normal
	escaped := false
	for _, char := range message {
		switch inputState {
		case SingleQuote:
			switch char {
			case '\'':
				inputState = Normal
			default:
				current = current + string(char)
			}
		case DoubleQuote:
			switch char {
			case '"':
				if escaped {
					current = current + string(char)
					escaped = false
				} else {
					inputState = Normal
				}
			case '\\':
				if escaped {
					current = current + string(char)
					escaped = false
				} else {
					escaped = true
				}
			default:
				if escaped {
					current = current + "\\" + string(char)
					escaped = false
				} else {
					current = current + string(char)
				}
			}
		default:
			switch char {
			case '\'':
				if escaped {
					current = current + string(char)
					escaped = false
				} else {
					inputState = SingleQuote
				}
			case '"':
				if escaped {
					current = current + string(char)
					escaped = false
				} else {
					inputState = DoubleQuote
				}
			case '\\':
				escaped = true
			case ' ':
				if escaped {
					current = current + string(char)
					escaped = false
				} else if current != "" {
					result = append(result, current)
					current = ""
				}
			default:
				current = current + string(char)
			}
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
