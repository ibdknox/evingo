package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func tree2dot(n node) string {
	count := 0
	nodes := make(map[node]string)
	result := "digraph foo {"
	translate := func(n node) string {
		var k string
		var ok bool
		if k, ok = nodes[n]; !ok {
			r := "n" + strconv.Itoa(count)
			count++
			nodes[n] = r
			return r
		}
		return k
	}
	translate(n)
	result += "}"
	return result
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		_ = strings.Split(line, " ")
	}
}
