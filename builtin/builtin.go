package builtin

import "fmt"
import "strings"
import "goober-lisp/base"
import "math"

type Builtin func([]base.Value) base.Value

func (v Builtin) Truthy() bool {
	return true
}

func (v Builtin) Prn() string {
	return "<builtin>"
}

func (v Builtin) String() string {
	return v.Prn()
}

var Builtins = map[string]Builtin{
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

func sexpr(v ...base.Value) base.Sexpr {
	return base.Sexpr(v)
}

func requireSymbol(v base.Value, msg string) base.Symbol {
	switch x := v.(type) {
	case base.Symbol:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireInt(v base.Value, msg string) base.Int {
	switch x := v.(type) {
	case base.Int:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireSexpr(v base.Value, msg string) base.Sexpr {
	switch x := v.(type) {
	case base.Sexpr:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireKeyword(v base.Value, msg string) base.Keyword {
	switch x := v.(type) {
	case base.Keyword:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func requireHashMap(v base.Value, msg string) HashMap {
	switch x := v.(type) {
	case HashMap:
		return x
	default:
		panic(fmt.Sprintf(msg+": %v", v))
	}
}

func builtin_list(vals []base.Value) base.Value {
	return base.Sexpr(vals)
}

func builtin_first(vals []base.Value) base.Value {

	if len(vals) != 1 {
		panic(fmt.Sprintf("first takes only 1 parameter: %v", vals))
	}

	seq := seq(vals[0])
	if len(seq) == 0 {
		return base.Nil{}
	} else {
		return seq[0]
	}
}

func builtin_rest(vals []base.Value) base.Value {

	if len(vals) != 1 {
		panic(fmt.Sprintf("rest takes only 1 parameter: %v", vals))
	}

	list := requireSexpr(vals[0], "rest takes a list")

	if len(list) == 0 {
		return base.Nil{}
	} else {
		return base.Sexpr(list[1:]) // this will probably bite me in the ass
	}
}

func builtin_cons(vals []base.Value) base.Value {

	if len(vals) != 2 {
		panic(fmt.Sprintf("cons takes only 2 parameters: %v", vals))
	}

	x := vals[0]
	list := requireSexpr(vals[1], "second argument must be a list")

	newList := make([]base.Value, 0, len(list)+1)
	newList = append(newList, x)
	newList = append(newList, list...)

	return base.Sexpr(newList)
}

func builtin_count(vals []base.Value) base.Value {

	if len(vals) != 1 {
		panic(fmt.Sprintf("count takes only 1 parameters: %v", vals))
	}

	x := vals[0]

	switch x := x.(type) {
	case base.Sexpr:
		return base.Int(len(x))
	case HashMap:
		return base.Int(len(x))
	default:
		panic(fmt.Sprintf("count requires a collection: %v", vals))
	}
}

func builtin_println(vals []base.Value) base.Value {
	newList := make([]string, 0, len(vals))
	for _, v := range vals {
		newList = append(newList, v.Prn())
	}
	fmt.Println(strings.Join(newList, " "))
	return base.Nil{}
}

type HashMap map[base.Value]base.Value

func (v HashMap) Truthy() bool {
	return len(v) > 0
}

func (v HashMap) Prn() string {

	var items string

	if len(v) > 0 {
		kvs := make([]string, 0, len(v)+2)
		for k, val := range v {
			kvs = append(kvs, k.Prn(), val.Prn())
		}
		items = " " + strings.Join(kvs, " ")
	}

	return "(hash-map" + items + ")"
}

func (v HashMap) String() string {
	return v.Prn()
}

func builtin_hashmap(vals []base.Value) base.Value {

	if math.Mod(float64(len(vals)), 2) != 0 {
		panic(fmt.Sprintf("hash-map's arguments must be an even number of values: %v", vals))
	}

	kvs := map[base.Value]base.Value{}
	for i := 0; i < len(vals); i += 2 {
		k := vals[i]
		val := vals[i+1]
		kvs[k] = val
	}

	return HashMap(kvs)
}

func builtin_get(vals []base.Value) base.Value {

	if len(vals) != 2 {
		panic(fmt.Sprintf("get takes 2 parameters: %v", vals))
	}

	m := requireHashMap(vals[0], "first argument must be a map")

	return m[vals[1]]
}

func builtin_put(vals []base.Value) base.Value {

	if len(vals) != 3 {
		panic(fmt.Sprintf("get takes 3 parameters: %v", vals))
	}

	m := requireHashMap(vals[0], "first argument must be a map")

	copy := make(map[base.Value]base.Value, len(m))
	for k, v := range m {
		copy[k] = v
	}
	copy[vals[1]] = vals[2]

	return HashMap(copy)
}

func seq(val base.Value) base.Sexpr {
	switch val := val.(type) {
	case HashMap:
		seq := make([]base.Value, 0, len(val)*2)
		for k, v := range val {
			seq = append(seq, base.Sexpr([]base.Value{k, v}))
		}
		return base.Sexpr(seq)
	case base.Sexpr:
		return val
	default:
		panic(fmt.Sprintf("not seq-able: %v", val))
	}
}

func builtin_seq(vals []base.Value) base.Value {

	if len(vals) != 1 {
		panic(fmt.Sprintf("seq takes 1 parameter: %v", vals))
	}

	return seq(vals[0])
}

func builtin_plus(vals []base.Value) base.Value {
	var x int = 0
	for i := range vals {
		val := requireInt(vals[i], "arguments to '+' must be numbers")
		x = x + int(val)
	}
	return base.Int(x)
}

func builtin_eq(vals []base.Value) base.Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("= takes at least 1 parameter: %v", vals))
	}

	x := int(requireInt(vals[0], "arguments to '=' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '=' must be numbers"))
		if val != x {
			return base.Boolean(false)
		}
	}
	return base.Boolean(true)
}

func builtin_lt(vals []base.Value) base.Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("< takes at least 1 parameter: %v", vals))
	}

	x := int(requireInt(vals[0], "arguments to '<' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '<' must be numbers"))
		if x >= val {
			return base.Boolean(false)
		}
		x = val
	}
	return base.Boolean(true)
}

func builtin_lteq(vals []base.Value) base.Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("<= takes at least 1 parameter: %v", vals))
	}

	x := int(requireInt(vals[0], "arguments to '<=' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '<=' must be numbers"))
		if x > val {
			return base.Boolean(false)
		}
		x = val
	}
	return base.Boolean(true)
}

func builtin_gt(vals []base.Value) base.Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf("> takes at least 1 parameter: %v", vals))
	}

	x := int(requireInt(vals[0], "arguments to '>' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '>' must be numbers"))
		if x <= val {
			return base.Boolean(false)
		}
	}
	return base.Boolean(true)
}

func builtin_gteq(vals []base.Value) base.Value {

	if len(vals) < 1 {
		panic(fmt.Sprintf(">= takes at least 1 parameter: %v", vals))
	}

	x := int(requireInt(vals[0], "arguments to '>=' must be numbers"))
	for _, i := range vals[1:] {
		val := int(requireInt(i, "arguments to '>=' must be numbers"))
		if x < val {
			return base.Boolean(false)
		}
	}
	return base.Boolean(true)
}
