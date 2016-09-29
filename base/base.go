package base

import "strconv"
import "strings"

// Defines the basic union of types that can be used
// as parameters or return values.
type Value interface {
	Truthy() bool
	Prn() string
}

// incorporate the reader data structures as values

type Nil struct{}
type Boolean bool
type Symbol string
type Int int
type Str string
type Sexpr []Value
type Keyword string

func (v Nil) Truthy() bool {
	return false
}

func (v Nil) Prn() string {
	return "nil"
}

func (v Nil) String() string {
	return v.Prn()
}

func (v Boolean) Truthy() bool {
	return bool(v)
}

func (v Boolean) Prn() string {
	return strconv.FormatBool(bool(v))
}

func (v Symbol) Truthy() bool {
	return true
}

func (v Symbol) Prn() string {
	return string(v)
}

func (v Int) Truthy() bool {
	return int(v) != 0
}

func (v Int) Prn() string {
	return strconv.Itoa(int(v))
}

func (v Str) Truthy() bool {
	trimmed := strings.TrimSpace(string(v))
	return len(trimmed) > 0
}

func (v Str) Prn() string {
	return string(v)
}

func (v Sexpr) Truthy() bool {
	return true
}

func (v Sexpr) Prn() string {
	list := []Value(v)

	elements := make([]string, 0, len(list))
	for _, i := range list {
		elements = append(elements, i.Prn())
	}

	return "(" + strings.Join(elements, " ") + ")"
}

func (v Sexpr) String() string {
	return v.Prn()
}

func (v Keyword) Prn() string {
	return ":" + string(v)
}

func (v Keyword) Truthy() bool {
	return true
}

func (v Keyword) String() string {
	return v.Prn()
}
