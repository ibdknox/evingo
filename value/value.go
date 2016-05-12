package value

import (
	"github.com/witheve/evingo/decimal"
)

type Value interface {
	Equal(Value) bool
	Hash() uint32
	String() string
}

type Uuid struct {
	top    uint32
	bottom uint64
}

func (u Uuid) Equal(v Value) bool {
	if u2, ok := v.(Uuid); ok {
		return (u.top == u2.top) && (u.bottom == u2.bottom)
	}
	return false
}

func (u Uuid) Hash() uint32 {
	return 0
}

func (u Uuid) String() string {
	return ""
}

type Text struct {
	s string
}

func (t Text) Equal(v Value) bool {
	if t2, ok := v.(Text); ok {
		return t.s == t2.s
	}
	return false
}

func (t Text) Hash() uint32 {
	return 0
}

func (t Text) String() string {
	return t.s
}

func NewText(s string) Value {
	return &Text{s}
}

type Number struct {
	d decimal.Decimal
}

func (n Number) Equal(v Value) bool {
	if t2, ok := v.(Number); ok {
		return n.d == t2.d
	}
	return false
}

func (n Number) Hash() uint32 {
	return 0
}

func (n Number) String() string {
	return n.d.String()
}

func NewNumberFromFloat(n float64) Value {
	return &Number{decimal.NewFromFloat(n)}
}
func NewNumberFromInt(n int64) Value {
	return &Number{decimal.New(n, 1)}
}
func NewNumberFromString(n string) Value {
	var d, err = decimal.NewFromString(n)
	if err != nil {
		panic("Invalid decimal string: " + n)
	}
	return &Number{d}
}

type Boolean struct {
	b bool
}

func (b Boolean) Equal(v Value) bool {
	if b2, ok := v.(Boolean); ok {
		return b.b == b2.b
	}
	return false

}

func (b Boolean) Hash() uint32 {
	if b.b {
		return 1
	}
	return 0
}

func (b Boolean) String() string {
	if b.b {
		return "true"
	}
	return "false"
}

func NewBoolean(b bool) Value {
	return &Boolean{b}
}
