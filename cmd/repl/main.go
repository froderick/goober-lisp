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

func handle(ns *goober.Ns, input string) {

	// not supposed to panic across packages, but too bad
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%s: %s", e, debug.Stack())
		}
	}()

	if !isEmpty(input) {
		for _, val := range goober.Read(input) {
			fmt.Printf("%v\n", goober.Eval(ns, val))
		}
	}
}

func main() {
	ns := goober.DefaultNs()
	for {

		reader := bufio.NewReader(os.Stdin)
		fmt.Print(ns.Name + "> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		handle(ns, input)
	}
}
