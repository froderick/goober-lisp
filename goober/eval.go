package goober

import "math"
import "fmt"
import "strings"

// incorporate functions as value types

type fn struct {
	args    []Symbol
	exprs   []Value
	context context
}

type recur []Value

func (v fn) truthy() bool {
	return true
}

func (v fn) prn() string {

	args := make([]string, 0, len(v.args))
	for _, arg := range v.args {
		args = append(args, arg.prn())
	}

	exprs := make([]string, 0, len(v.exprs))
	for _, expr := range v.exprs {
		exprs = append(exprs, expr.prn())
	}

	return "(fn (" + strings.Join(args, " ") + ") " + strings.Join(args, " ") + ")"
}

func (v fn) String() string {
	return v.prn()
}

func (v recur) truthy() bool {
	return true
}

func (v recur) prn() string {
	return fmt.Sprintf("#recur[%v]", v)
}

// data structures to support vars and bindings

type Ns struct {
	Name string
	vars map[string]Value
}

func NewNs(name string) Ns {
	return Ns{Name: "user", vars: map[string]Value{}}
}

func (ns *Ns) def(name string, value Value) {
	ns.vars[name] = value
}

func (ns *Ns) undef(name string) {
	delete(ns.vars, name)
}

type binding struct {
	name  string
	value Value
}

type context struct {
	ns       *Ns
	bindings []binding
}

func (c *context) push(name Symbol, value Value) {
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

func (c context) get(name Symbol) Value {

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

	if b, ok := builtinMap[string(s)]; ok {
		return b
	}

	panic("cannot find a binding or var with this symbol name: " + name)
}

// This function is special because it acts like a macro,
// it operates on the raw values haneded to it from the
// reader.
func special_def(context *context, vals []Value) Value {

	if len(vals) != 2 {
		panic(fmt.Sprintf("def takes only 2 parameters: %v", vals))
	}

	switch varname := vals[0].(type) {
	case Symbol:
		context.ns.def(string(varname), eval(context, vals[1]))
	default:
		panic(fmt.Sprintf("vars can only be named by symbols: %v", varname))
	}

	return Nil{}
}

func requireSymbol(v Value, msg string) Symbol {
	switch x := v.(type) {
	case Symbol:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireInt(v Value, msg string) Int {
	switch x := v.(type) {
	case Int:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireSexpr(v Value, msg string) Sexpr {
	switch x := v.(type) {
	case Sexpr:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireHashMap(v Value, msg string) HashMap {
	switch x := v.(type) {
	case HashMap:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireFn(v Value, msg string) fn {
	switch x := v.(type) {
	case fn:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func special_let(context *context, vals []Value) Value {

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

	var result Value
	for i := 1; i < len(vals); i++ {
		expr := vals[i]
		result = eval(context, expr)
	}

	// return the result of the last statement in the let block

	return result
}

func special_if(context *context, vals []Value) Value {

	if len(vals) < 2 {
		panic(fmt.Sprintf("if takes at least 2 parameters: %v", vals))
	}

	if len(vals) > 3 {
		panic(fmt.Sprintf("if takes at most 3 parameters: %v", vals))
	}

	test := eval(context, vals[0])

	switch v := test.(type) {
	case Value:
		if v.truthy() {
			return eval(context, vals[1])
		} else {
			if len(vals) == 2 {
				return Nil{}
			} else {
				return eval(context, vals[2])
			}
		}
	default:
		panic(fmt.Sprintf("only value types can be tested for truthiness: %v", test))
	}
}

func special_fn(context *context, vals []Value) fn {

	if len(vals) < 2 {
		panic(fmt.Sprintf("fn takes at least 2 parameters: %v", vals))
	}

	params := requireSexpr(vals[0], "expected args in the form of a list")

	names := make([]Symbol, 0, len(params))
	for i := range params {
		name := requireSymbol(params[i], "arguments to functions must be symbols")
		names = append(names, name)
	}

	// intentionally copying the context here, that becomes part of the fn
	return fn{args: names, exprs: vals[1:], context: *context}
}

func special_fn_call_inner(name string, fn *fn, context *context, vals []Value) Value {

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

	var result Value
	for _, expr := range fn.exprs {
		result = eval(&fn.context, expr)
	}

	return result
}

func special_fn_call(name string, fn fn, context *context, vals []Value) Value {

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

func special_do(context *context, vals []Value) Value {

	var result Value
	for _, expr := range vals {
		result = eval(context, expr)
	}

	return result
}

func special_quote(vals []Value) Value {
	if len(vals) != 1 {
		panic(fmt.Sprintf("quote takes only 1 parameter: %v", vals))
	}
	param := vals[0]
	return param
}

func builtin_recur(vals []Value) Value {
	return recur(vals)
}

func builtin_list(vals []Value) Value {
	return Sexpr(vals)
}

func builtin_first(vals []Value) Value {

	if len(vals) != 1 {
		panic(fmt.Sprintf("first takes only 1 parameter: %v", vals))
	}

	seq := seq(vals[0])
	if len(seq) == 0 {
		return Nil{}
	} else {
		return seq[0]
	}
}

func builtin_rest(vals []Value) Value {

	if len(vals) != 1 {
		panic(fmt.Sprintf("rest takes only 1 parameter: %v", vals))
	}

	list := requireSexpr(vals[0], "rest takes a list")

	if len(list) == 0 {
		return Nil{}
	} else {
		return Sexpr(list[1:]) // this will probably bite me in the ass
	}
}

func builtin_cons(vals []Value) Value {

	if len(vals) != 2 {
		panic(fmt.Sprintf("cons takes only 2 parameters: %v", vals))
	}

	x := vals[0]
	list := requireSexpr(vals[1], "second argument must be a list")

	newList := make([]Value, 0, len(list)+1)
	newList = append(newList, x)
	newList = append(newList, list...)

	return Sexpr(newList)
}

func builtin_count(vals []Value) Value {

	if len(vals) != 1 {
		panic(fmt.Sprintf("count takes only 1 parameters: %v", vals))
	}

	x := vals[0]

	switch x := x.(type) {
	case Sexpr:
		return Int(len(x))
	case HashMap:
		return Int(len(x))
	default:
		panic(fmt.Sprintf("count requires a collection: %v", vals))
	}
}

func builtin_println(vals []Value) Value {
	newList := make([]string, 0, len(vals))
	for _, v := range vals {
		newList = append(newList, v.prn())
	}
	fmt.Println(strings.Join(newList, " "))
	return Nil{}
}

// TODO NEXT: macros?

type HashMap map[Value]Value

func (v HashMap) truthy() bool {
	return len(v) > 0
}

func (v HashMap) prn() string {

	var items string

	if len(v) > 0 {
		kvs := make([]string, 0, len(v)+2)
		for k, val := range v {
			kvs = append(kvs, k.prn(), val.prn())
		}
		items = " " + strings.Join(kvs, " ")
	}

	return "(hash-map" + items + ")"
}

func (v HashMap) String() string {
	return v.prn()
}

func builtin_hashmap(vals []Value) Value {

	if math.Mod(float64(len(vals)), 2) != 0 {
		panic(fmt.Sprintf("hash-map's arguments must be an even number of values: %v", vals))
	}

	kvs := map[Value]Value{}
	for i := 0; i < len(vals); i += 2 {
		k := vals[i]
		val := vals[i+1]
		kvs[k] = val
	}

	return HashMap(kvs)
}

func builtin_get(vals []Value) Value {

	if len(vals) != 2 {
		panic(fmt.Sprintf("get takes 2 parameters: %v", vals))
	}

	m := requireHashMap(vals[0], "first argument must be a map")

	return m[vals[1]]
}

func builtin_put(vals []Value) Value {

	if len(vals) != 3 {
		panic(fmt.Sprintf("get takes 3 parameters: %v", vals))
	}

	m := requireHashMap(vals[0], "first argument must be a map")

	copy := make(map[Value]Value, len(m))
	for k, v := range m {
		copy[k] = v
	}
	copy[vals[1]] = vals[2]

	return HashMap(copy)
}

func seq(val Value) Sexpr {
	switch val := val.(type) {
	case HashMap:
		seq := make([]Value, 0, len(val)*2)
		for k, v := range val {
			seq = append(seq, Sexpr([]Value{k, v}))
		}
		return Sexpr(seq)
	case Sexpr:
		return val
	default:
		panic(fmt.Sprintf("not seq-able: %v", val))
	}
}

func builtin_seq(vals []Value) Value {

	if len(vals) != 1 {
		panic(fmt.Sprintf("seq takes 1 parameter: %v", vals))
	}

	return seq(vals[0])
}

func builtin_plus(vals []Value) Value {
	var base int = 0
	for i := range vals {
		val := requireInt(vals[i], "arguments to '+' must be numbers")
		base = base + int(val)
	}
	return Int(base)
}

func builtin_eq(vals []Value) Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("= takes at least 1 parameter: %v", vals))
	}

	base := int(requireInt(vals[0], "arguments to '=' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '=' must be numbers"))
		if val != base {
			return Boolean(false)
		}
	}
	return Boolean(true)
}

func builtin_lt(vals []Value) Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("< takes at least 1 parameter: %v", vals))
	}

	base := int(requireInt(vals[0], "arguments to '<' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '<' must be numbers"))
		if base >= val {
			return Boolean(false)
		}
		base = val
	}
	return Boolean(true)
}

func builtin_lteq(vals []Value) Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("<= takes at least 1 parameter: %v", vals))
	}

	base := int(requireInt(vals[0], "arguments to '<=' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '<=' must be numbers"))
		if base > val {
			return Boolean(false)
		}
		base = val
	}
	return Boolean(true)
}

func builtin_gt(vals []Value) Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("> takes at least 1 parameter: %v", vals))
	}

	base := int(requireInt(vals[0], "arguments to '>' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '>' must be numbers"))
		if base <= val {
			return Boolean(false)
		}
	}
	return Boolean(true)
}

func builtin_gteq(vals []Value) Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf(">= takes at least 1 parameter: %v", vals))
	}

	base := int(requireInt(vals[0], "arguments to '>=' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '>=' must be numbers"))
		if base < val {
			return Boolean(false)
		}
	}
	return Boolean(true)
}

func evalRest(context *context, v Sexpr) []Value {
	rest := make([]Value, 0, len(v)-1)
	for _, item := range v[1:] {
		evaluated := eval(context, item)
		rest = append(rest, evaluated)
	}
	return rest
}

func evalSexpr(context *context, v Sexpr) Value {

	if len(v) == 0 {
		return v
	}

	switch first := v[0].(type) {

	case Sexpr:
		resolved := make([]Value, 0)
		resolved = append(resolved, eval(context, first))
		resolved = append(resolved, v[1:]...)
		return eval(context, Sexpr(resolved))

	case fn:
		return special_fn_call("anonymous", first, context, evalRest(context, v))

	case Keyword:

		if len(v) != 2 {
			panic(fmt.Sprintf("a keyword as a function takes only one argument: %v", v))
		}

		replacement := Sexpr([]Value{
			Symbol("get"),
			v[1],
			v[0],
		})

		return evalSexpr(context, replacement)

	case Symbol:

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
		}

		if builtin, ok := builtinMap[string(first)]; ok {
			return builtin(evalRest(context, v))
		}

		// bound functions

		resolved := context.get(first)

		switch resolved := resolved.(type) {
		case fn:
			return special_fn_call(string(first), resolved, context, evalRest(context, v))
		case builtin:
			return resolved(evalRest(context, v))
		case Keyword:

			// TODO: this is duplicated

			if len(v) != 2 {
				panic(fmt.Sprintf("a keyword as a function takes only one argument: %v", v))
			}

			replacement := Sexpr([]Value{
				Symbol("get"),
				v[1],
				resolved,
			})

			return evalSexpr(context, replacement)

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

type builtin func([]Value) Value

func (v builtin) truthy() bool {
	return true
}

func (v builtin) prn() string {
	return "<builtin>"
}

func (v builtin) String() string {
	return v.prn()
}

var builtinMap = map[string]builtin{
	"recur":    builtin_recur,
	"list":     builtin_list,
	"first":    builtin_first,
	"rest":     builtin_rest,
	"cons":     builtin_cons,
	"+":        builtin_plus,
	"=":        builtin_eq,
	">":        builtin_gt,
	">=":       builtin_gteq,
	"<":        builtin_lt,
	"<=":       builtin_lteq,
	"hash-map": builtin_hashmap,
	"get":      builtin_get,
	"put":      builtin_put,
	"seq":      builtin_seq,
	"println":  builtin_println,
	"count":    builtin_count,
}

// Evaluates a Value data structure as code.
func eval(context *context, v Value) Value {

	//fmt.Printf("eval() input: %v\n", v)
	//fmt.Printf("eval() bindings: %v\n", context.bindings)

	var result Value

	switch v := v.(type) {
	case Sexpr:
		result = evalSexpr(context, v)
	case Symbol:
		if "nil" == string(v) {
			return Nil{}
		}
		result = context.get(v)
	default:
		result = v
	}

	//fmt.Printf("eval() output: %v\n", result)

	return result
}

func Eval(ns *Ns, v Value) Value {

	context := context{
		ns:       ns,
		bindings: make([]binding, 0),
	}

	return eval(&context, v)
}
