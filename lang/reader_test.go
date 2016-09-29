package lang

import "testing"
import "fmt"
import "reflect"
import "goober-lisp/base"

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Log(fmt.Sprintf("%v and %v are not equal", a, b))
		t.Fail()
	}
}

func sexpr(v ...base.Value) base.Sexpr {
	return base.Sexpr(v)
}

func sym(s string) base.Symbol {
	return base.Symbol(s)
}

func readOne(s string) base.Value {
	return Read(s)[0]
}

func TestReadBoolean(t *testing.T) {
	assertEqual(t, readOne("true"), base.Boolean(true))
	assertEqual(t, readOne("false"), base.Boolean(false))
}

func TestReadSymbol(t *testing.T) {
	assertEqual(t, readOne("+"), readOne("+"))
}

func TestReadInt(t *testing.T) {
	assertEqual(t, readOne("100"), base.Int(100))
}

func TestReadStr(t *testing.T) {
	assertEqual(t, readOne("\"A\""), base.Str("A"))
}

func TestReadSexpr(t *testing.T) {
	assertEqual(t, readOne("(+ 1 2 3)"), sexpr(sym("+"), base.Int(1), base.Int(2), base.Int(3))) // basic
	assertEqual(t, readOne("(x (y))"), sexpr(sym("x"), sexpr(sym("y"))))                         // nesting
}

func TestReadQuote(t *testing.T) {
	assertEqual(t, readOne("'(foo)"), sexpr(sym("quote"), sexpr(sym("foo"))))
	assertEqual(t, readOne("'f"), sexpr(sym("quote"), (sym("f"))))
}

func doPop(ts TokenStream) string {
	x, _ := ts.Pop()
	return x
}

func TestTokenStream(t *testing.T) {
	ts := NewTokenStream("a", "b", "c")
	assertEqual(t, []string{doPop(ts), doPop(ts), doPop(ts)}, []string{"a", "b", "c"})
}
