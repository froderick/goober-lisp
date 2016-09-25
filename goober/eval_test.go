package goober

import "testing"

func test_eval(s string) Value {
	ns := NewNs("user")
	sexpr := Read(s)
	return Eval(&ns, sexpr)
}

func TestEvalPlus(t *testing.T) {
	assertEqual(t, test_eval("(+ 1 2 3)"), Int(6))
}

func TestEvalDef(t *testing.T) {
	ns := NewNs("user")
	assertEqual(t, Eval(&ns, Read("(def f \"F\")")), Nil{})
	assertEqual(t, Eval(&ns, Read("f")), Str("F"))
}

func TestEvalLet(t *testing.T) {
	assertEqual(t, test_eval("(let (a 1 b 2) a)"), Int(1))
}

func TestEvalQuote(t *testing.T) {
	assertEqual(t, test_eval("'y"), Symbol("y"))
}

func TestEvalIf(t *testing.T) {

	assertEqual(t, test_eval("(if nil 'y 'n)"), Symbol("n"))

	assertEqual(t, test_eval("(if true 'y 'n)"), Symbol("y"))
	assertEqual(t, test_eval("(if false 'y 'n)"), Symbol("n"))

	assertEqual(t, test_eval("(if 'x 'y 'n)"), Symbol("y"))

	assertEqual(t, test_eval("(if -1 'y 'n)"), Symbol("y"))
	assertEqual(t, test_eval("(if 0 'y 'n)"), Symbol("n"))
	assertEqual(t, test_eval("(if 1 'y 'n)"), Symbol("y"))

	assertEqual(t, test_eval("(if \"\" 'y 'n)"), Symbol("n"))
	// assertEqual(t, test_eval("(if \" \" 'y 'n)"), Symbol("y")) // TODO: this will not work because our lexer sucks
	assertEqual(t, test_eval("(if \"test\" 'y 'n)"), Symbol("y"))

	assertEqual(t, test_eval("(if () 'y 'n)"), Symbol("y"))
	assertEqual(t, test_eval("(if '(1) 'y 'n)"), Symbol("y"))

}
