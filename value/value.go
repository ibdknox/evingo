package value

import (
	"github.com/witheve/evingo/decimal"
)

type Value interface {
	Equal(Value) bool
	Hash() uint32
	String() string
}

type uuid struct {
	top    uint32
	bottom uint64
}

func (u uuid) Equal(v Value) bool {
	if u2, ok := v.(uuid); ok {
		return (u.top == u2.top) && (u.bottom == u2.bottom)
	}
	return false
}

func (u uuid) Hash() uint32 {
	return 0
}

func (u uuid) String() string {
	return ""
}

type text struct {
	s string
}

func (t text) Equal(v Value) bool {
	if t2, ok := v.(text); ok {
		return t.s == t2.s
	}
	return false
}

func (t text) Hash() uint32 {
	return 0
}

func (t text) String() string {
	return ""
}

type number struct {
	d decimal.Decimal
}

func (n number) Equal(v Value) bool {
	if t2, ok := v.(number); ok {
		return n.d == t2.d
	}
	return false
}

func (n number) Hash() uint32 {
	return 0
}

func (n number) String() string {
	return n.d.String()
}

type boolean struct {
	b bool
}

func (b boolean) Equal(v Value) bool {
	if b2, ok := v.(boolean); ok {
		return b.b == b2.b
	}
	return false

}

func (n boolean) Hash() uint32 {
	if n.b {
		return 1
	}
	return 0
}

func (n boolean) String() string {
	if n.b {
		return "true"
	}
	return "false"

}
