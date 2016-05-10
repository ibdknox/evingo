package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func tree2dot(node n) {
    count := 0
	nodes := map[node]string
    result := "digraph foo {"
    translate := fn (n node) string {
        if k, ok := nodes[n] ; !ok {
            r := "n" + count
            count++
            nodes[n] = r
            return r
        }
    }
    k := n.children[]
    result += "}"
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		terms := strings.Split(line, " ")
	}
}
