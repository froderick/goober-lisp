package lang

import "testing"
import "fmt"
import "runtime/debug"
import "reflect"
import "goober-lisp/base"

// eval utilities

func test_eval_ns(ns *Ns, s string) base.Value {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%s: %s", e, debug.Stack())
		}
	}()
	sexpr := Read(s)[0]
	return Eval(ns, sexpr)
}

func test_eval(s string) base.Value {
	ns := DefaultNs()
	return test_eval_ns(ns, s)
}

// eval test data

type xform func(base.Value) base.Value

type testPair struct {
	input    string
	expected base.Value
	xform    xform
}

func pair(input string, expected base.Value) testPair {
	return testPair{input: input, expected: expected}
}

var tests = []testPair{

	// special functions (macros)

	testPair{
		input: `(let (def-result (def x 100))
	              (list def-result x))`,
		expected: sexpr(base.Nil{}, base.Int(100)),
	},

	testPair{input: "(let (a 1 b 2) a)", expected: base.Int(1)},

	testPair{input: "(if nil 'y 'n)", expected: base.Symbol("n")},
	testPair{input: "(if true 'y 'n)", expected: base.Symbol("y")},
	testPair{input: "(if false 'y 'n)", expected: base.Symbol("n")},
	testPair{input: "(if 'x 'y 'n)", expected: base.Symbol("y")},
	testPair{input: "(if -1 'y 'n)", expected: base.Symbol("y")},
	testPair{input: "(if 0 'y 'n)", expected: base.Symbol("n")},
	testPair{input: "(if 1 'y 'n)", expected: base.Symbol("y")},
	testPair{input: "(if \"\" 'y 'n)", expected: base.Symbol("n")},
	//testPair{input: "(if \" \" 'y 'n)", expected: Symbol("y")}, // TODO: this will not work because our lexer sucks
	testPair{input: "(if \"test\" 'y 'n)", expected: base.Symbol("y")},
	testPair{input: "(if () 'y 'n)", expected: base.Symbol("y")},
	testPair{input: "(if '(1) 'y 'n)", expected: base.Symbol("y")},

	testPair{
		input: "(fn (a) (+ 1 2) (+ a 10))",
		expected: fn{
			args: []base.Symbol{base.Symbol("a")},
			exprs: sexpr(
				sexpr(base.Symbol("+"), base.Int(1), base.Int(2)),
				sexpr(base.Symbol("+"), base.Symbol("a"), base.Int(10)),
			),
		},
		xform: func(v base.Value) base.Value { // empty out the context so the data structures match
			fn := v.(fn)
			fn.context = context{}
			return fn
		},
	},
	testPair{input: "((fn (a) (+ 1 2) (+ a 10)) 5)", expected: base.Int(15)},

	testPair{input: "'y", expected: base.Symbol("y")},
	testPair{input: "'(1 2 3)", expected: sexpr(base.Int(1), base.Int(2), base.Int(3))},
	testPair{input: "(quote (1 2 3))", expected: sexpr(base.Int(1), base.Int(2), base.Int(3))},
	testPair{input: "(not 'x)", expected: base.Boolean(false)},

	testPair{input: "(do (+ 1 2 3) 5)", expected: base.Int(5)},

	// builtin functions (not macros)

	testPair{input: "(list 1 2 3)", expected: sexpr(base.Int(1), base.Int(2), base.Int(3))},
	testPair{input: "(first '(1 2 3))", expected: base.Int(1)},
	testPair{input: "(rest '(1 2 3))", expected: sexpr(base.Int(2), base.Int(3))},
	testPair{input: "(cons 100 '())", expected: sexpr(base.Int(100))},
	testPair{input: "(+ 1 2 3)", expected: base.Int(6)},

	// keywords as basic functions

	testPair{input: "(:a (hash-map :a :B))", expected: base.Keyword("B")},

	// keywords as higher-order functions
	testPair{
		input:    "(map :a (list (hash-map :a \"ONE\") (hash-map :a \"TWO\")))",
		expected: sexpr(base.Str("ONE"), base.Str("TWO")),
	},
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
