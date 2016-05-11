package value

import (
	"bytes"
	"strconv"
)

type Node interface {
	String() string
	Children() []attribute
	Lookup(string) (Node, bool)
}

type attribute struct {
	name  string
	value Node
}

type Mapnode struct {
	m map[string]Node
}

func (m Mapnode) String() string {
	result := "{"
	for k, v := range m.m {
		result += k + "," + v.String()
	}
	result += "}"
	return result
}

func (m Mapnode) Children() []attribute {
	var result []attribute
	for k, v := range m.m {
		result = append(result, attribute{k, v})
	}
	return result
}

func (m Mapnode) Lookup(key string) (Node, bool) {
	v, ok := m.m[key]
	if ok {
		return v, true
	}
	return nil, false
}

type Setnode struct {
	m []Node
}

func (m Setnode) String() string {
	result := "{"
	for _, k := range m.m {
		result += k.String() + ","
	}
	result += "}"
	return result
}

func (m Setnode) Children() []attribute {
	var result []attribute
	for _, k := range m.m {
		result = append(result, attribute{"", k})
	}
	return result
}

//does this have a lookup
func (m Setnode) Lookup(key string) (Node, bool) {
	return nil, false
}

type Valnode struct {
	v Value
}

func NewValnode(v Value) Node {
	return &Valnode{v}
}

func (v Valnode) String() string {
	return v.String()
}

func (v Valnode) Children() []attribute {
	return nil
}

func (v Valnode) Lookup(key string) (Node, bool) {
	return nil, false
}

func Insert(n Node, path []string, value Node) {
	h := n.(*Mapnode)
	for _, i := range path[:(len(path) - 1)] {
		z, ok := h.Lookup(i)
		if !ok {
			m := &Mapnode{make(map[string]Node)}
			h.m[i] = m
			h = m
		} else {
			h = z.(*Mapnode)
		}
	}
	h.m[path[len(path)-1]] = value
}

func lookup(n Node, path []string) (Node, bool) {
	var ok bool
	for _, i := range path {
		n, ok = n.Lookup(i)
		if !ok {
			return nil, false
		}
	}
	return n, true
}

func Tree2dot(n Node) string {
	count := 0
	nodes := make(map[Node]string)
	var result bytes.Buffer
	result.WriteString("digraph foo {\n")
	var translate func(n Node) string
	translate = func(n Node) string {
		var k string
		var ok bool
		if k, ok = nodes[n]; !ok {
			k = "n" + strconv.Itoa(count)
			count++
			switch n.(type) {
			case *Mapnode, *Setnode:
				nodes[n] = k
				for _, v := range n.Children() {
					result.WriteString("  " + k + "->" + translate(v.value) + " [label=\"" + v.name + "\"]\n")
				}
			default:
				result.WriteString("  " + k + " [label=\"" + n.String() + "\"]\n")
			}
		}
		return k
	}
	translate(n)
	result.WriteString("}\n")
	return result.String()
}

func NewMapNode() Node {
	return &Mapnode{make(map[string]Node)}
}
