package lang

import "strconv"
import "strings"
import "errors"
import "goober-lisp/base"

// Split an s-expression string into tokens
func tokenize(s string) []string {

	s = strings.Replace(s, ",", "", -1)
	s = strings.Replace(s, "(", " ( ", -1)
	s = strings.Replace(s, ")", " ) ", -1)
	s = strings.Replace(s, "'", " ' ", -1)
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
func parseAtom(s string) base.Value {

	if "true" == s {
		return base.Boolean(true)
	}
	if "false" == s {
		return base.Boolean(false)
	}

	if ival, err := strconv.Atoi(s); err == nil {
		return base.Int(ival)
	} else if len(s) > 1 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		i := s[1 : len(s)-1]
		return base.Str(i)
	} else if strings.HasPrefix(s, ":") {
		return base.Keyword(s[1:])
	} else if len(s) > 0 {
		return base.Symbol(s)
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
// Returns a pointer to a base.Value.
func Parse(ts TokenStream) base.Value {

	token, err := ts.Pop()
	if err != nil {
		panic("no tokens supplied")
	}

	if token == "(" {
		elements := make([]base.Value, 0)
		for {
			if next, _ := ts.Peek(); next == ")" {
				ts.Pop() // dump )
				return base.Sexpr(elements)
			} else {
				v := Parse(ts)
				elements = append(elements, v)
			}
		}
	}

	val := parseAtom(token)

	switch val := val.(type) {
	case base.Symbol:
		if string(val) == "'" {
			return base.Sexpr([]base.Value{base.Symbol("quote"), Parse(ts)})
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

// The reader function to use when you want to read series of s-expressions in
// a string into base.Value data structures.
func Read(s string) []base.Value {
	tokens := tokenize(s)
	ts := NewTokenStream(tokens...)

	vals := make([]base.Value, 0)
	for {
		if _, err := ts.Peek(); err != nil {
			break // out of tokens
		}
		vals = append(vals, Parse(ts))
	}

	return vals
}
