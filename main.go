package main

import "fmt"
import "strings"
import "strconv"
import "os"
import "bufio"

/*

Core components of a lisp:

- reader (takes text, returns data structures)
- eval (takes data structures, interprets them as code and executes them)

*/

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
func read(s string) (v *Value) {
	tokens := tokenize(s)
	return parse(tokens)
}

// This function is special because it acts like a macro,
// it operates on the raw values haneded to it from the
// reader.
func special_def(env map[string]Value, vals []Value) Value {

	if len(vals) != 2 {
		panic("def takes only 2 parameters: " + Value{List: vals}.String())
	}

	varName := vals[0]
	if !varName.isSymbol() {
		panic("vars can only be named by symbols: " + varName.String())
	}

	varVal := vals[1]
	env[*varName.Symbol] = eval(env, varVal)
	return Value{}
}

func special_if(env map[string]Value, vals []Value) Value {

	if len(vals) < 2 {
		panic("if takes at least 2 parameters: " + Value{List: vals}.String())
	}

	if len(vals) > 3 {
		panic("if takes at most 3 parameters: " + Value{List: vals}.String())
	}

	test := eval(env, vals[0])

	if test.truthy() {
		return eval(env, vals[1])
	} else {
		if len(vals) == 2 {
			return Value{}
		} else {
			return eval(env, vals[2])
		}
	}
}

/* what does the execution context look like?
probably:
  - global vars state (this is a pointer, its mutable and everything can see it)
  - lexical bindings (this should be immutable for a given block of code)

  so we need to pass around this state, and use a function to resolve bindings from it

  we'll need a new Value type for functions. IFn? :)
*/

// TODO: the `env` map is really the `captured bindings` map. currently it reflects
// global vars, but it should probably actually be the lexical bindings currently
// in scope, with a fallback to global vars if nothing else matches.
func special_fn(env map[string]Value, vals []Value) Value {

	if len(vals) < 2 {
		panic("fn takes least 2 parameters: " + Value{List: vals}.String())
	}

	rawParams := vals[0]
	if !rawParams.isList() {
		panic("expected args in the form of a list: " + rawParams.String())
	}

	// eval the params to determine their bindings
	paramNames := make([]string, 0, len(rawParams.List))
	for i := range rawParams.List {

		name := rawParams.List[i]
		if !name.isSymbol() {
			panic("arguments to functions must be symbols: " + name.String())
		}

		paramNames = append(paramNames, *name.Symbol)
	}

	return Value{Fn: &Fn{Args: paramNames, Statements: vals[1:]}}
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

// Evaluates a Value data structure as code.
func eval(env map[string]Value, v Value) Value {

	if v.isList() {

		// empty lists are just empty lists
		if len(v.List) == 0 {
			return v
		}

		fn := v.List[0]
		if !fn.isSymbol() {
			fn = eval(env, fn)
		}
		if !fn.isSymbol() {
			panic("function names must be symbols: " + fn.String())
		}

		// special functions that take raw (un-eval'ed) arguments

		if "def" == *fn.Symbol {
			return special_def(env, v.List[1:])
		}

		if "if" == *fn.Symbol {
			return special_if(env, v.List[1:])
		}

		if "fn" == *fn.Symbol {
			return special_fn(env, v.List[1:])
		}

		// builtin functions

		args := make([]Value, 0, len(v.List)-1)
		for i := 1; i < len(v.List); i++ {
			elem := eval(env, v.List[i])
			args = append(args, elem)
		}

		if "+" == *fn.Symbol {
			return builtin_plus(args)
		}

		// defined functions

		if resolved, ok := env[*fn.Symbol]; ok {

			if !resolved.isFn() {
				panic("var is not bound to a function: " + *fn.Symbol)
			}

			// call fn here
			// - resolve lexical bindings in elements (perhaps do let first?)

		} else {
			panic("cannot find a var with this symbol name: " + *fn.Symbol)
		}

		panic("function not defined: " + fn.String())
	}

	if v.isSymbol() {
		if replace, ok := env[*v.Symbol]; ok {
			return replace
		} else {
			panic("cannot find a var with this symbol name: " + v.String())
		}
	}

	return v
}

func isEmpty(s string) bool {
	t := strings.TrimSpace(s)
	return len(t) == 0
}

func main() {
	env := make(map[string]Value)
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("goober-lisp> ")

		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if !isEmpty(text) {
			value := read(text)
			fmt.Printf("%+v\n", eval(env, *value))
		}
	}
}
