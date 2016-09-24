package goober

import "testing"

func test_eval(s string) Value {
	ns := ns{name: "user", vars: map[string]Value{}}
	sexpr := Read(s)
	return Eval(&ns, sexpr)
}

func TestEvalPlus(t *testing.T) {
	assertEqual(t, test_eval("(+ 1 2 3)"), Int(6))
}
