package goober

import "testing"
import "fmt"
import "runtime/debug"
import "reflect"

// eval utilities

func test_eval_ns(ns *Ns, s string) Value {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%s: %s", e, debug.Stack())
		}
	}()
	sexpr := Read(s)[0]
	return Eval(ns, sexpr)
}

func test_eval(s string) Value {
	ns := NewNs("user")
	return test_eval_ns(&ns, s)
}

// eval test data

type xform func(Value) Value

type testPair struct {
	input    string
	expected Value
	xform    xform
}

func pair(input string, expected Value) testPair {
	return testPair{input: input, expected: expected}
}

var tests = []testPair{

	// special functions (macros)

	testPair{
		input: `(let (def-result (def x 100))
	              (list def-result x))`,
		expected: sexpr(Nil{}, Int(100)),
	},

	testPair{input: "(let (a 1 b 2) a)", expected: Int(1)},

	testPair{input: "(if nil 'y 'n)", expected: Symbol("n")},
	testPair{input: "(if true 'y 'n)", expected: Symbol("y")},
	testPair{input: "(if false 'y 'n)", expected: Symbol("n")},
	testPair{input: "(if 'x 'y 'n)", expected: Symbol("y")},
	testPair{input: "(if -1 'y 'n)", expected: Symbol("y")},
	testPair{input: "(if 0 'y 'n)", expected: Symbol("n")},
	testPair{input: "(if 1 'y 'n)", expected: Symbol("y")},
	testPair{input: "(if \"\" 'y 'n)", expected: Symbol("n")},
	//testPair{input: "(if \" \" 'y 'n)", expected: Symbol("y")}, // TODO: this will not work because our lexer sucks
	testPair{input: "(if \"test\" 'y 'n)", expected: Symbol("y")},
	testPair{input: "(if () 'y 'n)", expected: Symbol("y")},
	testPair{input: "(if '(1) 'y 'n)", expected: Symbol("y")},

	testPair{
		input: "(fn (a) (+ 1 2) (+ a 10))",
		expected: fn{
			args: []Symbol{Symbol("a")},
			exprs: sexpr(
				sexpr(Symbol("+"), Int(1), Int(2)),
				sexpr(Symbol("+"), Symbol("a"), Int(10)),
			),
		},
		xform: func(v Value) Value { // empty out the context so the data structures match
			fn := v.(fn)
			fn.context = context{}
			return fn
		},
	},
	testPair{input: "((fn (a) (+ 1 2) (+ a 10)) 5)", expected: Int(15)},

	testPair{input: "'y", expected: Symbol("y")},
	testPair{input: "'(1 2 3)", expected: sexpr(Int(1), Int(2), Int(3))},
	testPair{input: "(quote (1 2 3))", expected: sexpr(Int(1), Int(2), Int(3))},

	testPair{input: "(do (+ 1 2 3) 5)", expected: Int(5)},

	// builtin functions (not macros)

	testPair{input: "(list 1 2 3)", expected: sexpr(Int(1), Int(2), Int(3))},
	testPair{input: "(first '(1 2 3))", expected: Int(1)},
	testPair{input: "(rest '(1 2 3))", expected: sexpr(Int(2), Int(3))},
	testPair{input: "(cons 100 '())", expected: sexpr(Int(100))},
	testPair{input: "(+ 1 2 3)", expected: Int(6)},
}

func TestEval(t *testing.T) {
	for _, pair := range tests {
		v := test_eval(pair.input)

		if pair.xform != nil {
			v = pair.xform(v)
		}

		if !reflect.DeepEqual(pair.expected, v) {
			t.Error(
				"For", pair.input,
				"expected", fmt.Sprintf("%v", pair.expected),
				"got", fmt.Sprintf("%v", v),
			)
		}
	}
}
