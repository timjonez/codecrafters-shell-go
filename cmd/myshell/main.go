package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

const SuccessCode = 0
const ErrCode = 1

type Mode string

const (
	Truncate Mode = ">"
	Append   Mode = ">>"
)

type FileDescriptor int

const (
	StdOut FileDescriptor = 1
	StdErr FileDescriptor = 2
)

type Redirect struct {
	Descriptor FileDescriptor
	Mode       Mode
}

type Input struct {
	Commands []string
	Redirect Redirect
	File     string
}

func (i *Input) HandleOut(stdout, stderr []byte) {
	if i.Redirect != (Redirect{}) {
		var flag int
		if i.Redirect.Mode == Append {
			flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		} else {
			flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		}

		file := strings.TrimSpace(i.File)
		f, err := os.OpenFile(file, flag, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			return
		}
		defer f.Close()

		if i.Redirect.Descriptor == StdOut {
			if _, err := f.Write(stdout); err != nil {
				stdout = append(stdout, []byte("\n")...)
				fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			}
		} else if i.Redirect.Descriptor == StdErr {
			if _, err := f.Write(stderr); err != nil {
				stderr = append(stderr, []byte("\n")...)
				fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			}
		}
	} else {
		if len(stdout) > 0 {
			fmt.Fprint(os.Stdout, string(stdout))
		}
		if len(stderr) > 0 {
			fmt.Fprint(os.Stderr, string(stderr))
		}
	}
}

func intoInput(message string) Input {
	file := ""
	redirect := Redirect{}
	if strings.Contains(message, ">>") {
		redirect = Redirect{Descriptor: StdOut, Mode: Append}
		message, file = splitInput(message, ">>")
	} else if strings.Contains(message, "1>") {
		redirect = Redirect{Descriptor: StdOut, Mode: Truncate}
		message, file = splitInput(message, "1>")
	} else if strings.Contains(message, ">") {
		redirect = Redirect{Descriptor: StdOut, Mode: Truncate}
		message, file = splitInput(message, ">")
	}
	commands := processCommands(message)

	return Input{
		Commands: commands,
		Redirect: redirect,
		File:     file,
	}
}

func splitInput(message, subStr string) (string, string) {
	parts := strings.Split(message, subStr)
	return parts[0], parts[1]
}

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		rawInput, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprint(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		rawInput = strings.TrimSpace(rawInput)
		input := intoInput(rawInput)
		commands := input.Commands

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
			input.HandleOut([]byte(cmd), []byte{})
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
			stdout, stderr, err := execFile(commands)
			if err != nil {
				fmt.Fprint(os.Stderr, "Error executing command")
			}
			input.HandleOut(stdout, stderr)
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

func execFile(args []string) ([]byte, []byte, error) {
	cmd := args[0]
	file := findFileOnPath(cmd)
	if file == "" {
		fmt.Fprint(os.Stderr, cmd+": command not found\n")
	}
	command := exec.Command(file, args[1:]...)

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return stdout.Bytes(), stderr.Bytes(), err
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

type InputType int

const (
	Normal InputType = iota
	SingleQuote
	DoubleQuote
)

func processCommands(message string) []string {
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
