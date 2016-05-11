package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func tree2dot(n node) string {
	count := 0
	nodes := make(map[node]string)
	var result bytes.Buffer
	result.WriteString("digraph foo {\n")
	var translate func(n node) string
	translate = func(n node) string {
		var k string
		var ok bool
		if k, ok = nodes[n]; !ok {
			k = "n" + strconv.Itoa(count)
			count++
			switch n.(type) {
			case *mapnode, *setnode:
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

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		_ = strings.Split(line, " ")
	}

	// genericize
	tree := &mapnode{make(map[string]node)}
	insert(tree, []string{"a", "b", "c"}, stringnode{"d"})
	insert(tree, []string{"a", "b", "z"}, stringnode{"e"})
	fmt.Println(tree2dot(tree))
}
