package main

import "testing"
import "fmt"
import "reflect"

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Log(fmt.Sprintf("%v and %v are not equal", a, b))
		t.Fail()
	}
}

func sexpr(v ...Value) Sexpr {
	return Sexpr(v)
}

func sym(s string) Symbol {
	return Symbol(s)
}

func TestReadBoolean(t *testing.T) {
	assertEqual(t, Read("true"), Boolean(true))
	assertEqual(t, Read("false"), Boolean(false))
}

func TestReadSymbol(t *testing.T) {
	assertEqual(t, Read("+"), Symbol("+"))
}

func TestReadInt(t *testing.T) {
	assertEqual(t, Read("100"), Int(100))
}

func TestReadStr(t *testing.T) {
	assertEqual(t, Read("\"A\""), Str("A"))
}

func TestReadSexpr(t *testing.T) {
	assertEqual(t, Read("(+ 1 2 3)"), sexpr(sym("+"), Int(1), Int(2), Int(3))) // basic
	assertEqual(t, Read("(x (y))"), sexpr(sym("x"), sexpr(sym("y"))))          // nesting
}

func TestReadQuote(t *testing.T) {
	assertEqual(t, Read("'(foo)"), sexpr(sym("quote"), sexpr(sym("foo"))))
}

func a(ts TokenStream) {
	ts.Pop()
}

func TestStringStream(t *testing.T) {
	var ts TokenStream = &StringStream{tokens: []string{"a", "b", "c"}}
	a(ts)
	x, _ := ts.Pop()
	assertEqual(t, x, "b")
}
