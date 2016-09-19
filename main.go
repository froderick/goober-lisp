package main

import "fmt"
import "strings"
import "errors"
import "strconv"

/*

Core components of a lisp:

- reader (takes text, returns data structures)
- eval (takes data structures, interprets them as code and executes them)

*/

type Value struct {
	Symbol *string
	Number *int
	Str    *string
	List   []Value
}

func toSexpr(v Value) string {
	if v.Symbol != nil {
		return *v.Symbol
	} else if v.Number != nil {
		return strconv.Itoa(*v.Number)
	} else if v.Str != nil {
		return "\"" + *v.Str + "\""
	} else if v.List != nil {

		elements := make([]string, 0)
		for i := range v.List {
			elements = append(elements, toSexpr(v.List[i]))
		}

		return "(" + strings.Join(elements, " ") + ")"
	} else {
		return "nil"
	}
}

func (v Value) String() string {
	return toSexpr(v)
}

func tokenize(s string) []string {

	s = strings.Replace(s, "(", " ( ", -1)
	s = strings.Replace(s, ")", " ) ", -1)
	parts := strings.Split(s, " ")

	tokens := make([]string, 0)
	for i := range parts {
		trimmed := strings.TrimSpace(parts[i])
		if len(trimmed) > 0 {
			tokens = append(tokens, trimmed)
		}
	}
	return tokens
}

func parseAtom(s string) *Value {

	//fmt.Println("parsing atom: " + s)

	ival, err := strconv.Atoi(s)
	if err == nil {
		return &Value{Number: &ival}
	}

	if len(s) > 1 {
		if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
			i := s[1 : len(s)-1]
			return &Value{Str: &i}
		} else {
			return &Value{Symbol: &s}
		}
	}

	return nil
}

func pop(items *[]string) *string {
	if len(*items) == 0 {
		return nil
	}
	first, rest := (*items)[0], (*items)[1:]
	*items = rest
	return &first
}

func _parse(tokens *[]string) (v *Value, err error) {

	//fmt.Printf("calling parse on tokens: %s\n", *tokens)

	if len(*tokens) == 0 {
		return nil, errors.New("no tokens supplied")
	}

	token := pop(tokens)

	if *token == "(" {
		sexpr := make([]Value, 0)
		for {
			if (*tokens)[0] == ")" {
				pop(tokens) // dump )
				return &Value{List: sexpr}, nil
			} else {
				v, err := _parse(tokens)
				if err != nil {
					return nil, errors.New("not a valid atom: " + *token)
				} else {
					sexpr = append(sexpr, *v)
				}
			}
		}
	}

	val := parseAtom(*token)
	if val == nil {
		return nil, errors.New("not a valid atom: " + *token)
	}
	return val, nil
}

func parse(tokens []string) (v *Value, err error) {
	stream := make([]string, len(tokens))
	copy(stream, tokens)
	return _parse(&stream)
}

func read(s string) (v *Value, err error) {
	tokens := tokenize(s)
	return parse(tokens)
}

func main() {

	value, err := read("(println \"hello\")")
	if err != nil {
		fmt.Println("error: " + err.Error())
	} else {
		fmt.Printf("%+v\n", value)
	}
}
