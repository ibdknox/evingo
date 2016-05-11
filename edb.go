package main

import (
	"github.com/witheve/evingo/cuckoo"
)

type edb struct {
	eav cuckoo.Cuckoo
}

type view struct {
	user uuid
	bag  uuid
	e    edb
}

func insert(v view, e, a, v Value) {
}

func scan_ea(v view, e, a) {
}

func allocate_bag(v view, e, a) {
}

