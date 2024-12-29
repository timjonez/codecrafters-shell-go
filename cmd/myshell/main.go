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
			cmd = strings.ReplaceAll(cmd, "'", "")
			fmt.Println(cmd)
		case "pwd":
			out, err := os.Getwd()
			if err != nil {
				fmt.Println("pwd: not found")
			}
			fmt.Println(out)
		case "type":
			cmd := commands[1]
			if slices.Contains([]string{"exit", "echo", "type", "pwd"}, cmd) {
				fmt.Println(cmd + " is a shell builtin")
			} else if handleTypeCommand(commands[1:]) {
				continue
			} else {
				fmt.Println(cmd + ": not found")
			}
		case "cd":
			cmd := commands[1]
			final := cmd
			if strings.Contains(final, "~") {
				final = strings.ReplaceAll(final, "~", os.Getenv("HOME"))
			}
			if err := os.Chdir(final); err != nil {
				fmt.Println("cd: " + cmd + ": No such file or directory")
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
		fmt.Println(cmd + ": command not found")
	}
	command := exec.Command(file, args[1:]...)
	output, err := command.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Print(string(output))
	return nil
}

func processInput(message string) []string {
	final := []string{}
	current := false
	currentIndex := 1
	args := strings.Fields(message)
	singleArgs := []string{}
	for _, arg := range strings.Split(message, "'") {
		if arg != " " {
			singleArgs = append(singleArgs, arg)
		}
	}
	for _, arg := range args {
		if strings.HasPrefix(arg, "'") {
			current = true
		} else if strings.HasSuffix(arg, "'") {
			final = append(final, strings.ReplaceAll(singleArgs[currentIndex], "'", ""))
			currentIndex += 1
			current = false
		} else if current {
		} else {
			final = append(final, arg)
		}
	}
	return final
}
