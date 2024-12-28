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
		commands := strings.Split(input, " ")

		switch commands[0] {
		case "exit":
			code, err := strconv.Atoi(commands[1])
			if err != nil {
				os.Exit(ErrCode)
			}
			os.Exit(code)
		case "echo":
			fmt.Println(strings.Join(commands[1:], " "))
		case "type":
			cmd := commands[1]
			if slices.Contains([]string{"exit", "echo", "type", "pwd"}, cmd) {
				fmt.Println(cmd + " is a shell builtin")
			} else if handleTypeCommand(commands[1:]) {
				continue
			} else {
				fmt.Println(cmd + ": not found")
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
