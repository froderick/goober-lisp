package main

import "fmt"
import "strconv"
import "strings"

type Value interface {
	truthy() bool
	prn() string
}

type Nil struct{}

func (v Nil) truthy() bool {
	return false
}

func (v Nil) prn() string {
	return "nil"
}

type Boolean bool

func (v Boolean) truthy() bool {
	return bool(v)
}

func (v Boolean) prn() string {
	return strconv.FormatBool(bool(v))
}

type Symbol string

func (v Symbol) truthy() bool {
	return true
}

func (v Symbol) prn() string {
	return string(v)
}

type Int int

func (v Int) truthy() bool {
	return int(v) != 0
}

func (v Int) prn() string {
	return strconv.Itoa(int(v))
}

type Str string

func (v Str) truthy() bool {
	trimmed := strings.TrimSpace(string(v))
	return len(trimmed) > 0
}

func (v Str) prn() string {
	return string(v)
}

type Sexpr []Value

func (v Sexpr) truthy() bool {
	return true
}

func (v Sexpr) prn() string {
	list := []Value(v)

	elements := make([]string, 0, len(list))
	for _, i := range list {
		elements = append(elements, i.prn())
	}

	return "(" + strings.Join(elements, " ") + ")"
}

func (v Sexpr) String() string {
	return v.prn()
}

// Split an s-expression string into tokens
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

// Parse an s-expression value as an atom, or return nil if no atom can be derived
func parseAtom(s string) Value {

	if "true" == s {
		return Boolean(true)
	}
	if "false" == s {
		return Boolean(false)
	}

	if ival, err := strconv.Atoi(s); err == nil {
		return Int(ival)
	} else if len(s) > 1 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		i := s[1 : len(s)-1]
		return Str(i)
	} else if len(s) > 0 {
		return Symbol(s)
	} else {
		panic("not a valid atom: " + s)
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

// The inner version of parse, takes a pointer to a slice of tokens.
// The slice is modified as the parsing logic consumes the tokens.
// Returns a pointer to a Value.
func _parse(tokens *[]string) Value {

	if len(*tokens) == 0 {
		panic("no tokens supplied")
	}

	token := pop(tokens)

	if *token == "(" {
		sexpr := make([]Value, 0)
		for {
			if (*tokens)[0] == ")" {
				pop(tokens) // dump )
				return Sexpr(sexpr)
			} else {
				v := _parse(tokens)
				sexpr = append(sexpr, v)
			}
		}
	}

	val := parseAtom(*token)

	switch val := val.(type) {
	case Symbol:
		if string(val) == "'" {
			return Sexpr([]Value{Symbol("quote"), _parse(tokens)})
		}
	default:
		return val
	}

	return val
}

// The public version of parse, takes a slice of tokens.
// Returns a pointer to a Value.
func parse(tokens []string) Value {
	stream := make([]string, len(tokens))
	copy(stream, tokens)
	return _parse(&stream)
}

// The reader function to use when you want to read an s-expression string
// into Value data structures.
func Read(s string) Value {
	tokens := tokenize(s)
	return parse(tokens)
}
