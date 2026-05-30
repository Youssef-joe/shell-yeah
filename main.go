package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	reader := bufio.NewReader(os.Stdin)
	builtinCommands := []string{"exit", "echo", "type"}

	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(0)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		inputParts := strings.Fields(input)
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
		default:
			if path, err := exec.LookPath(cmd); err == nil {
				command := exec.Command(path, args...)
				command.Args[0] = cmd
				command.Stdout = os.Stdout
				command.Stderr = os.Stderr
				command.Stdin = os.Stdin
				_ = command.Run()
			} else {
				fmt.Println(cmd + ": command not found")
			}
		}

	}
}

// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"time"
// )

// func boring(msg string) <-chan string {
// 	c := make(chan string)
// 	go func() {
// 		for i := 0; ;i++ {
// 			c <- fmt.Sprintf("%s %d", msg, i);
// 			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
// 		}
// 	}()
// 	return c;
// }

// func fanIn(input1, input2 <-chan string) <-chan string {
// 	c := make(chan string)
// 	go func() { for {c <- <-input1} } ()
// 	go func() { for {c <- <-input2} } ()
// 	return c
// }
// func fanIn2(input1, input2 <-chan string) <-chan string {
// 	c := make(chan string)
// 	go func() {
// 		for {
// 			select {
// 			case s := <-input1 : c <- s
// 			case s := <-input2: c <- s
// 			}
// 		}
// 	}()
// 	return c
// }

// func f(left, right chan int) {
// 	left <- 1 + <-right
// }
// func main() {
// 	// c := fanIn2(boring("joe"), boring("ann"))

// 	// for i := 0; i < 10; i++ {
// 	// 	fmt.Println(<-c)
// 	// }
// 	// fmt.Println("You're boring, I'm Leaving")

// 	// chinees whisper pattern, gopher style
// 	const n = 100000
// 	leftmost := make(chan int)
// 	right := leftmost
// 	left := leftmost

// 	for i := 0; i < n; i++ {
// 		right = make(chan int)
// 		go f(left, right)
// 		left = right
// 	}
// 	go func(c chan int) {c <- 1} (right)
// 	fmt.Println(<-leftmost)
// }
