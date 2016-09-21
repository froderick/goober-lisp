package goober

import "fmt"
import "strings"
import "strconv"

type Fn struct {
	Args       []string // the bindings we expect
	Statements []Value  // the statements we'll eval
}

type Value struct {
	Boolean *bool
	Symbol  *string
	Number  *int
	Str     *string
	List    []Value
	Fn      *Fn
}

func (v Value) isBoolean() bool {
	return v.Boolean != nil
}

func (v Value) isSymbol() bool {
	return v.Symbol != nil
}

func (v Value) isNumber() bool {
	return v.Number != nil
}

func (v Value) isStr() bool {
	return v.Str != nil
}

func (v Value) isList() bool {
	return v.List != nil
}

func (v Value) isFn() bool {
	return v.Fn != nil
}

func (v Value) isNil() bool {
	return v.Symbol == nil && v.Number == nil && v.Str == nil && v.List == nil
}

func (v Value) truthy() bool {
	if v.Boolean != nil {
		return *v.Boolean
	} else if v.Symbol != nil {
		return true
	} else if v.Number != nil {
		return true
	} else if v.Str != nil {
		trimmed := strings.TrimSpace(*v.Str)
		return len(trimmed) > 0
	} else if v.List != nil {
		return true
	} else if v.Fn != nil {
		return true
	} else {
		return false
	}
}

// Returns string s-expression to represent a Value.
func toSexpr(v Value) string {
	if v.Boolean != nil {
		return strconv.FormatBool(*v.Boolean)
	} else if v.Symbol != nil {
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
	} else if v.Fn != nil {
		return fmt.Sprintf("%+v\n", v.Fn)
	} else {
		return "nil"
	}
}

// Use toSexpr as a custom tostring formatter
func (v Value) String() string {
	return toSexpr(v)
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
func parseAtom(s string) *Value {

	if "true" == s {
		b := true
		return &Value{Boolean: &b}
	}
	if "false" == s {
		b := false
		return &Value{Boolean: &b}
	}

	ival, err := strconv.Atoi(s)
	if err == nil {
		return &Value{Number: &ival}
	} else if len(s) > 1 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
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

func sym(sym string) Value {
	return Value{Symbol: &sym}
}

func list(vals ...Value) Value {
	return Value{List: vals}
}

// The inner version of parse, takes a pointer to a slice of tokens.
// The slice is modified as the parsing logic consumes the tokens.
// Returns a pointer to a Value.
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

	if val.isSymbol() && *val.Symbol == "'" {
		quoted := _parse(tokens)
		q := list(sym("quote"), *quoted)
		return &q
	}

	return val
}

// The public version of parse, takes a slice of tokens.
// Returns a pointer to a Value.
func parse(tokens []string) (v *Value) {
	stream := make([]string, len(tokens))
	copy(stream, tokens)
	return _parse(&stream)
}

// The reader function to use when you want to read an s-expression string
// into Value data structures.
func Read(s string) (v *Value) {
	tokens := tokenize(s)
	return parse(tokens)
}
