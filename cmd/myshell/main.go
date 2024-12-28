package main

import (
	"bufio"
	"fmt"
	"os"
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
			if handleTypeCommand(commands[1:]) {
				continue
			} else if slices.Contains([]string{"exit", "echo", "type"}, cmd) {
				fmt.Println(cmd + " is a shell builtin")
			} else {
				fmt.Println(cmd + ": not found")
			}
		default:
			fmt.Println(commands[0] + ": command not found")
		}
	}
}

func handleTypeCommand(args []string) bool {
	cmd := args[0]
	dirs := strings.Split(os.Getenv("PATH"), ":")
	for _, dir := range dirs {
		file_path := dir + "/" + cmd
		_, err := os.Stat(file_path)
		if err == nil {
			fmt.Println(cmd + " is " + file_path)
			return true
		}
	}
	return false
}
