package goober

import "math"

//import "fmt"

type binding struct {
	Name  string
	Value Value
}

type context struct {
	globalVars map[string]Value
	bindings   []binding
}

func (c *context) def(name string, value Value) {
	c.globalVars[name] = value
}

func (c *context) push(name string, value Value) {
	b := binding{Name: name, Value: value}
	c.bindings = append(c.bindings, b)
}

func (c *context) pop() binding {

	if len(c.bindings) == 0 {
		panic("no bindings remain to be popped")
	}

	popped := c.bindings[len(c.bindings)-1]
	c.bindings = c.bindings[:len(c.bindings)-1]

	return popped
}

func (c context) get(name string) Value {

	if len(c.bindings) > 0 {
		for i := len(c.bindings); i > 0; i-- {
			binding := c.bindings[i-1]
			if name == binding.Name {
				return binding.Value
			}
		}
	}

	if globalVar, ok := c.globalVars[name]; ok {
		return globalVar
	}

	panic("cannot find a binding or var with this symbol name: " + name)
}

// This function is special because it acts like a macro,
// it operates on the raw values haneded to it from the
// reader.
func special_def(context *context, vals []Value) Value {

	if len(vals) != 2 {
		panic("def takes only 2 parameters: " + Value{List: vals}.String())
	}

	varName := vals[0]
	if !varName.isSymbol() {
		panic("vars can only be named by symbols: " + varName.String())
	}

	context.def(*varName.Symbol, eval(context, vals[1]))

	return Value{}
}

func special_let(context *context, vals []Value) Value {

	if len(vals) < 1 {
		panic("let takes at least 1 parameter: " + Value{List: vals}.String())
	}

	bindings := vals[0]
	if !bindings.isList() {
		panic("let's first argument must be a list: " + bindings.String())
	}

	if math.Mod(float64(len(bindings.List)), 2) != 0 {
		panic("let's binding list must be an even number of values: " + bindings.String())
	}

	pushes := 0

	// eval each binding, add it to the context
	for i := 0; i < len(bindings.List); i += 2 {

		bindingName := bindings.List[i]
		if !bindingName.isSymbol() {
			panic("bindings can only be made for symbols: " + bindingName.String())
		}

		bindingExpr := bindings.List[i+1]
		bindingValue := eval(context, bindingExpr)

		context.push(*bindingName.Symbol, bindingValue)

		pushes++
	}

	// eval the rest of the let arguments

	var result Value
	for i := 1; i < len(vals); i++ {
		expr := vals[i]
		result = eval(context, expr)
	}

	// pop off all the bindings

	for i := 0; i < pushes; i++ {
		context.pop() // TODO: these cleanups should happen even if evaulation fails
	}

	// return the result of the last statement in the let block

	return result
}

func special_if(context *context, vals []Value) Value {

	if len(vals) < 2 {
		panic("if takes at least 2 parameters: " + Value{List: vals}.String())
	}

	if len(vals) > 3 {
		panic("if takes at most 3 parameters: " + Value{List: vals}.String())
	}

	test := eval(context, vals[0])

	if test.truthy() {
		return eval(context, vals[1])
	} else {
		if len(vals) == 2 {
			return Value{}
		} else {
			return eval(context, vals[2])
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
func special_fn(vals []Value) Value {

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
func eval(context *context, v Value) Value {

	if v.isList() {

		// empty lists are just empty lists
		if len(v.List) == 0 {
			return v
		}

		fn := v.List[0]
		if !fn.isSymbol() {
			fn = eval(context, fn)
		}
		if !fn.isSymbol() {
			panic("function names must be symbols: " + fn.String())
		}

		// special functions that take raw (un-eval'ed) arguments

		if "def" == *fn.Symbol {
			return special_def(context, v.List[1:])
		}

		if "let" == *fn.Symbol {
			return special_let(context, v.List[1:])
		}

		if "if" == *fn.Symbol {
			return special_if(context, v.List[1:])
		}

		if "fn" == *fn.Symbol {
			return special_fn(v.List[1:])
		}

		// builtin functions

		args := make([]Value, 0, len(v.List)-1)
		for i := 1; i < len(v.List); i++ {
			elem := eval(context, v.List[i])
			args = append(args, elem)
		}

		if "+" == *fn.Symbol {
			return builtin_plus(args)
		}

		// defined functions

		resolved := context.get(*fn.Symbol)

		if !resolved.isFn() {
			panic("symbol is not bound to a function: " + *fn.Symbol)
		}

		// call fn here
		// - resolve lexical bindings in elements (perhaps do let first?)

		panic("function not defined: " + fn.String())
	}

	if v.isSymbol() {
		return context.get(*v.Symbol)
	}

	return v
}

func Eval(globalVars map[string]Value, v Value) Value {

	context := context{
		globalVars: globalVars,
		bindings:   make([]binding, 0),
	}

	return eval(&context, v)
}