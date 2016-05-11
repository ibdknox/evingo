package main

import (
	"github.com/witheve/evingo/cuckoo"
	"github.com/witheve/evingo/value"
)

type edb struct {
	eav cuckoo.Cuckoo
}

type context struct {
	user value.Uuid
	bag  value.Uuid
	// time restriction
	e edb
}

func insert(c context, e, a, v value.Value) {
}

func scan_ea(c context, e, a value.Value) {
}

func allocate_bag(c context, e, a value.Value) {
}
