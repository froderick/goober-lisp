package main

import "fmt"
import "os"
import "bufio"
import "strings"
import "goober-lisp/goober"
import "runtime/debug"

func isEmpty(s string) bool {
	t := strings.TrimSpace(s)
	return len(t) == 0
}

func handle(globalVars map[string]goober.Value, input string) {

	// not supposed to panic across packages, but too bad
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%s: %s", e, debug.Stack())
		}
	}()

	if !isEmpty(input) {
		value := goober.Read(input)
		fmt.Printf("%v\n", goober.Eval(globalVars, *value))
	}
}

func main() {
	globalVars := make(map[string]goober.Value)
	for {

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("goober-lisp> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		handle(globalVars, input)
	}
}
