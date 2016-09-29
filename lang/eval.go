package lang

import "math"
import "fmt"
import "strings"
import "goober-lisp/base"
import "goober-lisp/builtin"

// incorporate functions as value types

type fn struct {
	args    []base.Symbol
	exprs   []base.Value
	context context
}

type recur []base.Value

func (v fn) Truthy() bool {
	return true
}

func (v fn) Prn() string {

	args := make([]string, 0, len(v.args))
	for _, arg := range v.args {
		args = append(args, arg.Prn())
	}

	exprs := make([]string, 0, len(v.exprs))
	for _, expr := range v.exprs {
		exprs = append(exprs, expr.Prn())
	}

	return "(fn (" + strings.Join(args, " ") + ") " + strings.Join(args, " ") + ")"
}

func (v fn) String() string {
	return v.Prn()
}

func (v recur) Truthy() bool {
	return true
}

func (v recur) Prn() string {
	return fmt.Sprintf("#recur[%v]", v)
}

// data structures to support vars and bindings

type Ns struct {
	Name string
	vars map[string]base.Value
}

func NewNs(name string) Ns {
	return Ns{Name: "user", vars: map[string]base.Value{}}
}

func (ns *Ns) def(name string, value base.Value) {
	ns.vars[name] = value
}

func (ns *Ns) undef(name string) {
	delete(ns.vars, name)
}

type binding struct {
	name  string
	value base.Value
}

type context struct {
	ns       *Ns
	bindings []binding
}

func (c *context) push(name base.Symbol, value base.Value) {
	b := binding{name: string(name), value: value}
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

func (c context) get(name base.Symbol) base.Value {

	s := string(name)

	if len(c.bindings) > 0 {
		for i := len(c.bindings); i > 0; i-- {
			binding := c.bindings[i-1]
			if s == binding.name {
				return binding.value
			}
		}
	}

	if v, ok := c.ns.vars[s]; ok {
		return v
	}

	if b, ok := builtin.Builtins[string(s)]; ok {
		return b
	}

	panic("cannot find a binding or var with this symbol name: " + name)
}

// This function is special because it acts like a macro,
// it operates on the raw values haneded to it from the
// reader.
func special_def(context *context, vals []base.Value) base.Value {

	if len(vals) != 2 {
		panic(fmt.Sprintf("def takes only 2 parameters: %v", vals))
	}

	switch varname := vals[0].(type) {
	case base.Symbol:
		context.ns.def(string(varname), eval(context, vals[1]))
	default:
		panic(fmt.Sprintf("vars can only be named by symbols: %v", varname))
	}

	return base.Nil{}
}

func requireSexpr(v base.Value, msg string) base.Sexpr {
	switch x := v.(type) {
	case base.Sexpr:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireSymbol(v base.Value, msg string) base.Symbol {
	switch x := v.(type) {
	case base.Symbol:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func special_let(context *context, vals []base.Value) base.Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("let takes at least 1 parameter: %v", vals))
	}

	bindings := requireSexpr(vals[0], "vars can only be named by symbols")

	if math.Mod(float64(len(bindings)), 2) != 0 {
		panic(fmt.Sprintf("let's binding list must be an even number of values: %v", bindings))
	}

	pushes := 0

	// eval each binding, add it to the context
	for i := 0; i < len(bindings); i += 2 {

		sym := requireSymbol(bindings[i], "bindings can only be made for symbols")
		expr := bindings[i+1]
		val := eval(context, expr)

		context.push(sym, val)
		defer context.pop()

		pushes++
	}

	// eval the rest of the let arguments

	var result base.Value
	for i := 1; i < len(vals); i++ {
		expr := vals[i]
		result = eval(context, expr)
	}

	// return the result of the last statement in the let block

	return result
}

func special_if(context *context, vals []base.Value) base.Value {

	if len(vals) < 2 {
		panic(fmt.Sprintf("if takes at least 2 parameters: %v", vals))
	}

	if len(vals) > 3 {
		panic(fmt.Sprintf("if takes at most 3 parameters: %v", vals))
	}

	test := eval(context, vals[0])

	switch v := test.(type) {
	case base.Value:
		if v.Truthy() {
			return eval(context, vals[1])
		} else {
			if len(vals) == 2 {
				return base.Nil{}
			} else {
				return eval(context, vals[2])
			}
		}
	default:
		panic(fmt.Sprintf("only value types can be tested for truthiness: %v", test))
	}
}

type IFn interface {
	invoke(name string, context *context, args []base.Value) base.Value
}

func special_fn(context *context, vals []base.Value) fn {

	if len(vals) < 2 {
		panic(fmt.Sprintf("fn takes at least 2 parameters: %v", vals))
	}

	params := requireSexpr(vals[0], "expected args in the form of a list")

	names := make([]base.Symbol, 0, len(params))
	for i := range params {
		name := requireSymbol(params[i], "arguments to functions must be symbols")
		names = append(names, name)
	}

	// intentionally copying the context here, that becomes part of the fn
	return fn{args: names, exprs: vals[1:], context: *context}
}

func special_fn_call_inner(name string, fn *fn, context *context, vals []base.Value) base.Value {

	//fmt.Printf("calling: %v(%v)\n", name, vals)

	if len(vals) < len(fn.args) {
		panic(fmt.Sprintf("%v takes %v parameters: %v", name, len(fn.args), vals))
	}

	for i, bindingname := range fn.args {
		bindingExpr := vals[i]
		bindingValue := bindingExpr
		fn.context.push(bindingname, bindingValue)
		defer fn.context.pop()
	}

	var result base.Value
	for _, expr := range fn.exprs {
		result = eval(&fn.context, expr)
	}

	return result
}

func special_fn_call(name string, fn fn, context *context, vals []base.Value) base.Value {

	result := special_fn_call_inner(name, &fn, context, vals)
	for {
		switch r := result.(type) {
		case recur:
			result = special_fn_call_inner(name, &fn, context, r)
		default:
			return result
		}
	}
}

func special_keyword_call(context *context, k base.Keyword, args []base.Value) base.Value {

	if len(args) != 1 {
		panic(fmt.Sprintf("a keyword as a function takes only one argument: %v", args))
	}

	replacement := base.Sexpr([]base.Value{
		base.Symbol("get"),
		args[0],
		k,
	})

	return evalSexpr(context, replacement)
}

func special_do(context *context, vals []base.Value) base.Value {

	var result base.Value
	for _, expr := range vals {
		result = eval(context, expr)
	}

	return result
}

func special_quote(vals []base.Value) base.Value {
	if len(vals) != 1 {
		panic(fmt.Sprintf("quote takes only 1 parameter: %v", vals))
	}
	param := vals[0]
	return param
}

func special_recur(vals []base.Value) base.Value {
	return recur(vals)
}

// TODO NEXT: macros?

func evalRest(context *context, v base.Sexpr) []base.Value {
	rest := make([]base.Value, 0, len(v)-1)
	for _, item := range v[1:] {
		evaluated := eval(context, item)
		rest = append(rest, evaluated)
	}
	return rest
}

func evalSexpr(context *context, v base.Sexpr) base.Value {

	if len(v) == 0 {
		return v
	}

	switch first := v[0].(type) {

	case base.Sexpr:
		resolved := make([]base.Value, 0)
		resolved = append(resolved, eval(context, first))
		resolved = append(resolved, v[1:]...)
		return eval(context, base.Sexpr(resolved))

	case fn:
		return special_fn_call("anonymous", first, context, evalRest(context, v))

	case base.Keyword:
		return special_keyword_call(context, first, evalRest(context, v))

	case base.Symbol:

		// special functions

		rawArgs := v[1:]

		switch first {
		case "def":
			return special_def(context, rawArgs)
		case "let":
			return special_let(context, rawArgs)
		case "if":
			return special_if(context, rawArgs)
		case "fn":
			return special_fn(context, rawArgs)
		case "quote":
			return special_quote(rawArgs)
		case "do":
			return special_do(context, rawArgs)
		case "recur":
			return special_recur(evalRest(context, v))
		}

		// builtin functions

		if builtin, ok := builtin.Builtins[string(first)]; ok {
			return builtin(evalRest(context, v))
		}

		// functions bound to symbols, vars, or things we permit to be used as functions (keywords)

		resolved := context.get(first)

		switch resolved := resolved.(type) {
		case fn:
			return special_fn_call(string(first), resolved, context, evalRest(context, v))
		case builtin.Builtin:
			return resolved(evalRest(context, v))
		case base.Keyword:
			return special_keyword_call(context, resolved, evalRest(context, v))

		default:

			// TODO: consider defining IFn interface so that we can return
			// different types of things that support being invoked as a
			// first-class function here (like :keywords)

			panic(fmt.Sprintf("symbol must be bound to a function: %v", resolved))
		}
	default:
		panic(fmt.Sprintf("not a valid function: %", first))
	}
}

// Evaluates a base.Value data structure as code.
func eval(context *context, v base.Value) base.Value {

	//fmt.Printf("eval() input: %v\n", v)
	//fmt.Printf("eval() bindings: %v\n", context.bindings)

	var result base.Value

	switch v := v.(type) {
	case base.Sexpr:
		result = evalSexpr(context, v)
	case base.Symbol:
		if "nil" == string(v) {
			return base.Nil{}
		}
		result = context.get(v)
	default:
		result = v
	}

	//fmt.Printf("eval() output: %v\n", result)

	return result
}

func Eval(ns *Ns, v base.Value) base.Value {

	context := context{
		ns:       ns,
		bindings: make([]binding, 0),
	}

	return eval(&context, v)
}
