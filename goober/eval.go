package goober

import "math"
import "fmt"

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

func special_fn(vals []Value) Value {

	if len(vals) < 2 {
		panic("fn takes least 2 parameters: " + Value{List: vals}.String())
	}

	rawParams := vals[0]
	if !rawParams.isList() {
		panic("expected args in the form of a list: " + rawParams.String())
	}

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

func special_fn_call(name string, fn Fn, context *context, vals []Value) Value {

	if len(vals) < len(fn.Args) {
		panic(fmt.Sprintf("%v takes %v parameters: %v", name, len(fn.Args), Value{List: vals}.String()))
	}

	for i, bindingName := range fn.Args {
		bindingExpr := vals[i]
		bindingValue := eval(context, bindingExpr)
		context.push(bindingName, bindingValue)
	}

	var result Value
	for _, expr := range fn.Statements { // TODO: these are not statements, they are expressions
		result = eval(context, expr)
	}

	for range fn.Args {
		context.pop() // TODO: these cleanups should happen even if evaulation fails
	}

	return result
}

func special_quote(vals []Value) Value {
	if len(vals) != 1 {
		panic(fmt.Sprintf("quote takes only 1 parameter: %v", Value{List: vals}))
	}
	param := vals[0]
	return param
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

func evalArgs(context *context, v Value) []Value {
	evaluatedArgs := make([]Value, 0, len(v.List)-1)
	for i := 1; i < len(v.List); i++ {
		elem := eval(context, v.List[i])
		evaluatedArgs = append(evaluatedArgs, elem)
	}
	return evaluatedArgs
}

// Evaluates a Value data structure as code.
func eval(context *context, v Value) Value {

	if v.isList() {

		if len(v.List) == 0 {
			return v
		}

		fn := v.List[0]

		if fn.isList() {
			fn = eval(context, fn)
		}

		if fn.isFn() {
			return special_fn_call("anonymous", *fn.Fn, context, evalArgs(context, v))
		}

		if fn.isSymbol() {

			sym := *fn.Symbol

			// special functions

			rawArgs := v.List[1:]

			switch sym {
			case "def":
				return special_def(context, rawArgs)
			case "let":
				return special_let(context, rawArgs)
			case "if":
				return special_if(context, rawArgs)
			case "fn":
				return special_fn(rawArgs)
			case "quote":
				return special_quote(rawArgs)
			case "+":
				return builtin_plus(evalArgs(context, v))
			}

			// bound functions

			resolved := context.get(sym)

			if !resolved.isFn() {
				panic("symbol is not bound to a function: " + sym)
			}

			return special_fn_call(sym, *resolved.Fn, context, evalArgs(context, v))
		}

		panic("function names must be symbols: " + fn.String())
	}

	if v.isSymbol() {

		if "nil" == *v.Symbol {
			return Value{}
		}

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
