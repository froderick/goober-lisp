package main

import "fmt"

func main() {

	v := []string{"a"}

	for i := len(v); i > 0; i-- {
		fmt.Printf("%v\n", i)
		//fmt.Println(v[i])
	}
}
