package main

import (
	"github.com/witheve/evingo/gotomic"
	"github.com/witheve/evingo/value"
)

type edb struct {
	h         *gotomic.Hash
	listeners map[*func(e, a, v value.Value)]struct{}
}

type attributeSet struct {
	h         *gotomic.Hash
	listeners map[*func(a, v value.Value)]struct{}
}

func NewAttributeSet() *attributeSet {
	return nil
}

type valueSet struct {
	h         *gotomic.Hash
	listeners map[*func(v value.Value)]struct{}
}

func NewValueSet() *valueSet {
	return nil
}

type context struct {
	user value.Uuid
	bag  value.Uuid
	// time restriction
	e edb
}

// per bag
func NewEdb() *edb {
	return &edb{
		h:         gotomic.NewHash(),
		listeners: make(map[*func(e, a, v value.Value)]struct{}),
	}
}

func insert(c context, e, a, v value.Value) {
	// there is a race here that we can close
	// by refactoring the interface, but the consequence
	// of losing it is only a pointless allocation
	var as interface{}
	var vs interface{}
	var ok bool

	if as, ok = c.e.h.Get(e); !ok {
		as = c.e.h.PutIfMissing(e, NewAttributeSet())
	}
	if vs, ok = as.(*attributeSet).h.Get(a); !ok {
		vs = as.(*attributeSet).h.PutIfMissing(a, NewValueSet())
	}
	vs.(*valueSet).h.Put(v, struct{}{})
}

func scan_ea(c context, e, a value.Value) {
}

func allocate_bag(c context, e, a value.Value) *value.Uuid {
	return nil
}
