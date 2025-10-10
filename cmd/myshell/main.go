package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
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
	if len(stdout) > 0 && !bytes.HasSuffix(stdout, []byte("\n")) {
		stdout = append(stdout, '\n')
	}

	if len(stderr) > 0 && !bytes.HasSuffix(stderr, []byte("\n")) {
		stderr = append(stderr, []byte("\n")...)
	}

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

		if i.Redirect.Descriptor == StdErr {
			if _, err := f.Write(stderr); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			}
			fmt.Fprint(os.Stdout, string(stdout))
		} else if i.Redirect.Descriptor == StdOut {
			if _, err := f.Write(stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			}
			fmt.Fprint(os.Stderr, string(stderr))
		}
	} else {
		if len(stderr) > 0 {
			fmt.Fprint(os.Stderr, string(stderr))
		}
		if len(stdout) > 0 {
			fmt.Fprint(os.Stdout, string(stdout))
		}
	}
}

func intoInput(message string) Input {
	file := ""
	redirect := Redirect{}
	if strings.Contains(message, "1>>") {
		redirect = Redirect{Descriptor: StdOut, Mode: Append}
		message, file = splitInput(message, "1>>")
	} else if strings.Contains(message, "2>>") {
		redirect = Redirect{Descriptor: StdErr, Mode: Append}
		message, file = splitInput(message, "2>>")
	} else if strings.Contains(message, ">>") {
		redirect = Redirect{Descriptor: StdOut, Mode: Append}
		message, file = splitInput(message, ">>")
	} else if strings.Contains(message, "1>") {
		redirect = Redirect{Descriptor: StdOut, Mode: Truncate}
		message, file = splitInput(message, "1>")
	} else if strings.Contains(message, "2>") {
		redirect = Redirect{Descriptor: StdErr, Mode: Truncate}
		message, file = splitInput(message, "2>")
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

type CustomCompleter struct {
	Completer readline.AutoCompleter
	Terminal  *readline.Terminal
	Log       *Log
}

// https://pkg.go.dev/github.com/chzyer/readline#PrefixCompleterInterface
func (c *CustomCompleter) Print(prefix string, level int, buf *bytes.Buffer) {
  fmt.Println("--------->", prefix)
  buf.WriteString("Hello world")
}

func (c *CustomCompleter) Do(line []rune, pos int) ([][]rune, int) {
	newline, length := c.Completer.Do(line, pos)
	if len(newline) == 0 && c.Log.LastKey != '\t' {
		c.Terminal.Bell()
		return newline, length
	}
	dirs := strings.Split(os.Getenv("PATH"), ":")
	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			info, err := os.Stat(fmt.Sprintf("%s/%s", dir, file.Name()))
			if err != nil {
				continue
			}

			if !file.IsDir() && info.Mode()&0111 == 0 {
				continue
			}
			if strings.Contains(file.Name(), string(line)) {
				completion := strings.Replace(file.Name(), string(line), "", 1)
				newline = append(newline, []rune(fmt.Sprintf("%s ", completion)))
			}
		}
	}
	if len(newline) == 0 {
		c.Terminal.Bell()
	}
	return newline, length
}

type Log struct {
	Current rune
	LastKey rune
}

func (l *Log) LogInput(key rune) (rune, bool) {
	l.LastKey = l.Current
	l.Current = key
	return key, true
}

type CustomWriter struct {
	writer io.Writer
}

func (w *CustomWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "\n") { // Detect completion output
		res := []string{"\n"}
		cmds := strings.Split(string(p), " ")
		for _, cmd := range cmds {
			if cmd == " " {
				continue
			}
			res = append(res, strings.TrimSpace(cmd))
		}
		return w.writer.Write([]byte(strings.Join(res, "  ")))
	}
	return w.writer.Write(p)
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("echo"),
	readline.PcItem("exit"),
)

func main() {
	log := Log{}

	customWriter := &CustomWriter{writer: readline.Stdout}

	config := readline.Config{
		Prompt: "$ ",
		AutoComplete: &CustomCompleter{
			Completer: completer,
			Log:       &log,
		},
		FuncFilterInputRune: log.LogInput,
		Stdout:              customWriter,
	}
	rl, err := readline.NewEx(&config)
	if err != nil {
		os.Exit(1)
	}
	defer rl.Close()

	config.AutoComplete.(*CustomCompleter).Terminal = rl.Terminal

	for {
		line, err := rl.Readline()
		if err != nil {
			fmt.Fprint(os.Stderr, "Error reading input:", err, "\n")
			os.Exit(1)
		}

		rawInput := strings.TrimSpace(line)
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
				fmt.Fprint(os.Stderr, "pwd: not found", "\n")
			}
			fmt.Println(out)
		case "type":
			cmd := commands[1]
			if slices.Contains([]string{"exit", "echo", "type", "pwd"}, cmd) {
				fmt.Fprint(os.Stdout, cmd+" is a shell builtin", "\n")
			} else if handleTypeCommand(commands[1:]) {
				continue
			} else {
				fmt.Fprint(os.Stderr, cmd+": not found\n")
			}
		case "cd":
			cmd := commands[1]
			final := cmd
			if strings.Contains(final, "~") {
				final = strings.ReplaceAll(final, "~", os.Getenv("HOME"))
			}
			if err := os.Chdir(final); err != nil {
				fmt.Fprint(os.Stderr, "cd: "+cmd+": No such file or directory", "\n")
			}
		default:
			stdout, stderr, err := execFile(commands)
			if err != nil {
				fmt.Fprint(os.Stderr, "Error executing command", "\n")
			}
			input.HandleOut(stdout, stderr)
		}
	}
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	// Check if any executable bit is set (user, group, or others)
	return info.Mode()&0111 != 0
}

func findFileOnPath(name string) string {
	dirs := strings.Split(os.Getenv("PATH"), ":")
	for _, dir := range dirs {
		file_path := dir + "/" + name
		if isExecutable(file_path) {
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
		_, err := os.Stat(cmd)
		if err != nil {
			fmt.Fprint(os.Stderr, cmd+": command not found\n")
		}
		file = cmd
	}

	command := exec.Command(file, args[1:]...)
	command.Args[0] = cmd
	// command := exec.Command(file, args[1:]...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return stdout.Bytes(), stderr.Bytes(), nil
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
