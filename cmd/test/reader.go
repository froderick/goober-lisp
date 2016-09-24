package main

import "strconv"
import "strings"
import "errors"

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
type TokenStream interface {
	Peek() (string, error)
	Pop() (string, error)
}

// The inner version of parse, takes a pointer to a slice of tokens.
// The slice is modified as the parsing logic consumes the tokens.
// Returns a pointer to a Value.
func Parse(ts TokenStream) Value {

	token, err := ts.Pop()
	if err != nil {
		panic("no tokens supplied")
	}

	if token == "(" {
		elements := make([]Value, 0)
		for {
			if next, _ := ts.Peek(); next == ")" {
				ts.Pop() // dump )
				return Sexpr(elements)
			} else {
				v := Parse(ts)
				elements = append(elements, v)
			}
		}
	}

	val := parseAtom(token)

	switch val := val.(type) {
	case Symbol:
		if string(val) == "'" {
			return Sexpr([]Value{Symbol("quote"), Parse(ts)})
		}
	default:
		return val
	}

	return val
}

type stringStream struct {
	tokens []string
}

func (s *stringStream) Peek() (string, error) {
	if len(s.tokens) == 0 {
		return "", errors.New("stream is empty")
	}
	first := s.tokens[0]
	return first, nil
}

func (s *stringStream) Pop() (string, error) {
	if len(s.tokens) == 0 {
		return "", errors.New("stream is empty")
	}
	first, rest := s.tokens[0], s.tokens[1:]
	s.tokens = rest
	return first, nil
}

func NewTokenStream(tokens ...string) TokenStream {
	return &stringStream{tokens: tokens}
}

// The reader function to use when you want to read an s-expression string
// into Value data structures.
func Read(s string) Value {
	tokens := tokenize(s)
	ts := NewTokenStream(tokens...)
	return Parse(ts)
}
