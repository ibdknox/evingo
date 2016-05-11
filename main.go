package main

import (
	"bufio"
	"fmt"
	"github.com/witheve/evingo/value"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		_ = strings.Split(line, " ")
	}

	// genericize
	tree := &value.mapnode{make(map[string]node)}
	value.insert(tree, []string{"a", "b", "c"}, value.stringnode{"d"})
	value.insert(tree, []string{"a", "b", "z"}, value.stringnode{"e"})
	fmt.Println(value.tree2dot(tree))
}
