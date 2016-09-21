package main

import "fmt"
import "os"
import "bufio"
import "strings"
import "goober-lisp/goober"

func isEmpty(s string) bool {
	t := strings.TrimSpace(s)
	return len(t) == 0
}

func main() {
	globalVars := make(map[string]goober.Value)
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("goober-lisp> ")

		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if !isEmpty(text) {
			value := goober.Read(text)
			fmt.Printf("%v\n", goober.Eval(globalVars, *value))
		}
	}
}
