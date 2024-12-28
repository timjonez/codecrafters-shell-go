package main

import (
	"bufio"
	"fmt"
	"os"
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
		default:
			fmt.Println(commands[0] + ": command not found")
		}
	}
}
