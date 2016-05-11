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
	tree := value.NewMapNode()
	value.Insert(tree, []string{"a", "b", "c"}, value.NewValnode(value.NewText("d")))
	value.Insert(tree, []string{"a", "b", "z"}, value.NewValnode(value.NewText("e")))
	fmt.Println(value.Tree2dot(tree))
}
