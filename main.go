package main

import "fmt"
import "strings"
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

func (v Value) isSymbol() bool {
	return v.Symbol != nil
}

func (v Value) isNumber() bool {
	return v.Number != nil
}

func toSexpr(v Value) string {
	if v.Symbol != nil {
		return *v.Symbol
	} else if v.Number != nil {
		return strconv.Itoa(*v.Number)
	} else if v.Str != nil {
		return "\"" + *v.Str + "\""
	} else if v.List != nil {

		elements := make([]string, 0, len(v.List))
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

	tokens := make([]string, 0, len(parts))
	for i := range parts {
		trimmed := strings.TrimSpace(parts[i])
		if len(trimmed) > 0 {
			tokens = append(tokens, trimmed)
		}
	}
	return tokens
}

func parseAtom(s string) *Value {

	ival, err := strconv.Atoi(s)
	if err == nil {
		return &Value{Number: &ival}
	}

	if len(s) > 1 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		i := s[1 : len(s)-1]
		return &Value{Str: &i}
	} else if len(s) > 0 {
		return &Value{Symbol: &s}
	} else {
		return nil
	}
}

// Useful for 'consuming' from the head of a slice. Returns the popped value
// from the slice, and updates the supplied slice pointer to not include the
// popped value.
func pop(items *[]string) *string {
	if len(*items) == 0 {
		return nil
	}
	first, rest := (*items)[0], (*items)[1:]
	*items = rest
	return &first
}

func _parse(tokens *[]string) (v *Value) {

	if len(*tokens) == 0 {
		panic("no tokens supplied")
	}

	token := pop(tokens)

	if *token == "(" {
		sexpr := make([]Value, 0)
		for {
			if (*tokens)[0] == ")" {
				pop(tokens) // dump )
				return &Value{List: sexpr}
			} else {
				v := _parse(tokens)
				sexpr = append(sexpr, *v)
			}
		}
	}

	val := parseAtom(*token)
	if val == nil {
		panic("not a valid atom: " + *token)
	}
	return val
}

func parse(tokens []string) (v *Value) {
	stream := make([]string, len(tokens))
	copy(stream, tokens)
	return _parse(&stream)
}

func read(s string) (v *Value) {
	tokens := tokenize(s)
	return parse(tokens)
}

func builtin_plus(vals []Value) Value {
	var base int = 0
	for i := range vals {
		val := vals[i]
		if !val.isNumber() {
			panic("arguments to '+' must be numbers: " + val.String())
		}
		base = base + *val.Number
	}
	return Value{Number: &base}
}

func eval(v Value) Value {

	if v.List == nil {
		return v
	}

	// empty lists are just empty lists
	if len(v.List) == 0 {
		return v
	}

	evaluated := make([]Value, 0, len(v.List))
	for i := range v.List {
		elem := eval(v.List[i])
		evaluated = append(evaluated, elem)
	}

	fn := evaluated[0]
	args := evaluated[1:]

	if !fn.isSymbol() {
		panic("function names must be symbols: " + fn.String())
	}

	if "+" == *fn.Symbol {
		return builtin_plus(args)
	}

	panic("function not defined: " + fn.String())
}

func main() {
	value := read("(+ 1 2 3)")
	fmt.Printf("%+v\n", eval(*value))
}
