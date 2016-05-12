package value

import (
	"github.com/witheve/evingo/decimal"
)

// we dont close over the value to avoid taking the closure, although
// its not that bad here
type writer func(Value, []byte, int)

type Value interface {
	Equals(interface{}) bool
	HashCode() uint32
	String() string
	Serialize() (int, writer)
	Deserialize([]byte, int)
}

type Uuid struct {
	top    uint32
	bottom uint64
}

func (u Uuid) Serialize() (int, writer) {
	return 12, func(v Value, target []byte, offset int) {
	}
}

func (u Uuid) Deserialize(source []byte, offset int) {
}

func (u Uuid) Equals(v interface{}) bool {
	if u2, ok := v.(Uuid); ok {
		return (u.top == u2.top) && (u.bottom == u2.bottom)
	}
	return false
}

func (u Uuid) HashCode() uint32 {
	return 0
}

func (u Uuid) String() string {
	return ""
}

type Text struct {
	s string
}

func (t Text) Equals(v interface{}) bool {
	if t2, ok := v.(Text); ok {
		return t.s == t2.s
	}
	return false
}

func (t Text) HashCode() uint32 {
	return 0
}

func (t Text) String() string {
	return t.s
}

func (t Text) Serialize() (int, writer) {
	return 12, func(v Value, target []byte, offset int) {
	}
}

func (t Text) Deserialize(source []byte, offset int) {
}

func NewText(s string) Value {
	return &Text{s}
}

type Number struct {
	d decimal.Decimal
}

func (n Number) Equals(v interface{}) bool {
	if t2, ok := v.(Number); ok {
		return n.d == t2.d
	}
	return false
}

func (n Number) HashCode() uint32 {
	return 0
}

func (n Number) String() string {
	return n.d.String()
}

func (n Number) Deserialize(source []byte, offset int) {
}

func (n Number) Serialize() (int, writer) {
	return 12, func(v Value, target []byte, offset int) {
	}
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

func (b Boolean) Equals(v interface{}) bool {
	if b2, ok := v.(Boolean); ok {
		return b.b == b2.b
	}
	return false

}

func (b Boolean) HashCode() uint32 {
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

func (b Boolean) Deserialize(source []byte, offset int) {
}

func (b Boolean) Serialize() (int, writer) {
	return 12, func(v Value, target []byte, offset int) {
	}
}
