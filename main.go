package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// fmt is kept like a potted plant: present for aesthetics and occasional prints.
var _ = fmt.Print

func printBanner() {
	green   := "\x1b[32;1m"
	cyan    := "\x1b[36;1m"
	magenta := "\x1b[35;1m"
	reset   := "\x1b[0m"

	fmt.Println(green + "════════════════════════════════════════════════════════════" + reset)
	fmt.Println(magenta + "  ███████╗██╗  ██╗███████╗██╗     ██╗      " + reset)
	fmt.Println(magenta + "  ██╔════╝██║  ██║██╔════╝██║     ██║      " + reset)
	fmt.Println(magenta + "  ███████╗███████║█████╗  ██║     ██║      " + reset)
	fmt.Println(magenta + "  ╚════██║██╔══██║██╔══╝  ██║     ██║      " + reset)
	fmt.Println(magenta + "  ███████║██║  ██║███████╗███████╗███████╗ " + reset)
	fmt.Println(magenta + "  ╚══════╝╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝ " + reset)
	fmt.Println(cyan    + "  ██╗   ██╗███████╗ █████╗ ██╗  ██╗██╗     " + reset)
	fmt.Println(cyan    + "  ╚██╗ ██╔╝██╔════╝██╔══██╗██║  ██║██║     " + reset)
	fmt.Println(cyan    + "   ╚████╔╝ █████╗  ███████║███████║██║     " + reset)
	fmt.Println(cyan    + "    ╚██╔╝  ██╔══╝  ██╔══██║██╔══██║╚═╝     " + reset)
	fmt.Println(cyan    + "     ██║   ███████╗██║  ██║██║  ██║██╗     " + reset)
	fmt.Println(cyan    + "     ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝    " + reset)
	fmt.Println(green + "════════════════════════════════════════════════════════════" + reset)
	fmt.Println(cyan  + "  type 'help' for builtins · Ctrl-D to exit" + reset)
}

func main() {
	printBanner()
	reader := bufio.NewReader(os.Stdin)
	builtinCommands := []string{"exit", "echo", "type", "pwd", "cd"}

	for {
		user := os.Getenv("USER")
		cwd, _ := os.Getwd()
		base := filepath.Base(cwd)
		fmt.Printf("\x1b[1;32m%s\x1b[0m:\x1b[1;34m%s\x1b[0m$ ", user, base)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(0)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Trailing '&' means this command wants to party in the background.
		background := false
		if strings.HasSuffix(input, "&") {
			background = true
			input = strings.TrimSpace(strings.TrimSuffix(input, "&"))
		}

		// Pipelines: passing output like a relay race, left -> right.
		if strings.Contains(input, "|") {
			parts := strings.Split(input, "|")
			var cmds []*exec.Cmd
			for _, part := range parts {
				part = strings.TrimSpace(part)
				fields := strings.Fields(part)
				if len(fields) == 0 {
					continue
				}
				name := fields[0]
				args := fields[1:]

				// Builtins get stage fright and refuse to run inside pipelines.
				if slices.Contains(builtinCommands, name) {
					fmt.Fprintf(os.Stderr, "builtin %s not supported in pipeline\n", name)
					goto CONTINUE
				}

				path, err := exec.LookPath(name)
				if err != nil {
					fmt.Println(name + ": command not found")
					goto CONTINUE
				}

				c := exec.Command(path, args...)
				c.Args[0] = name
				c.Stderr = os.Stderr
				cmds = append(cmds, c)
			}

			// Hook up the plumbing: stdout -> pipe -> stdin.
			for i := 0; i < len(cmds)-1; i++ {
				r, w, _ := os.Pipe()
				cmds[i].Stdout = w
				cmds[i+1].Stdin = r
			}

			if len(cmds) > 0 {
				if cmds[0].Stdin == nil {
					cmds[0].Stdin = os.Stdin
				}
				if cmds[len(cmds)-1].Stdout == nil {
					cmds[len(cmds)-1].Stdout = os.Stdout
				}
			}

			// Kick off the performers (processes). Wish them luck.
			for _, c := range cmds {
				if err := c.Start(); err != nil {
					fmt.Fprintln(os.Stderr, "failed to start:", err)
				} else if background {
					if c.Process != nil {
						fmt.Printf("[%d] ", c.Process.Pid)
					}
				}
			}

			if !background {
				for _, c := range cmds {
					_ = c.Wait()
				}
			} else {
				fmt.Println()
			}

		CONTINUE:
			continue
		}

		inputParts := strings.Fields(input)
		if len(inputParts) == 0 {
			continue
		}

		cmd := inputParts[0]
		args := inputParts[1:]

		switch cmd {
		case "exit":
			os.Exit(0)
		case "echo":
			fmt.Println(strings.Join(args, " "))
		case "type":
			if len(args) == 0 {
				continue
			}
			if slices.Contains(builtinCommands, args[0]) {
				fmt.Println(args[0] + " is a shell builtin")
			} else if path, err := exec.LookPath(args[0]); err == nil {
				fmt.Println(args[0] + " is " + path)
			} else {
				fmt.Println(args[0] + ": not found")
			}
		case "pwd":
			if path, err := os.Getwd(); err == nil {
				fmt.Println(path)
			}
		case "cd":
			handleCd(args)
		default:
			if path, err := exec.LookPath(cmd); err == nil {
				command := exec.Command(path, args...)
				command.Args[0] = cmd
				command.Stdout = os.Stdout
				command.Stderr = os.Stderr
				command.Stdin = os.Stdin

				if background {
					if err := command.Start(); err != nil {
						fmt.Fprintln(os.Stderr, "failed to start background job:", err)
					} else if command.Process != nil {
						fmt.Printf("[%d]\n", command.Process.Pid)
					}
				} else {
					_ = command.Run()
				}

			} else {
				fmt.Println(cmd + ": command not found")
			}
		}

	}
}

func handleCd(args []string) {
	target := cdTarget(args)

	if err := os.Chdir(target); err != nil {
		fmt.Fprintf(os.Stdout, "cd: %s: No such file or directory\n", target)
	}
}

func cdTarget(args []string) string {
	if len(args) > 0 {
		return args[0]
	}

	return os.Getenv("HOME")
}

// Ancient experimental snippets have been banished to /dev/null.
// If you find a time-traveling gopher, offer it coffee and a ticket home.
