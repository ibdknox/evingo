package main

import (
	//"bufio"
	"fmt"
	"github.com/witheve/evingo/value"
	//"os"
	//"strings"
)

func main() {
	var s = "[[\"a\", \"attr1\", 7], [\"b\", \"attr1\", \"dog\"], [\"b\", \"attr2\", true]]"
	var b = []byte(s)
	fmt.Println(ReadFactsFromJson(b))

	//scanner := bufio.NewScanner(os.Stdin)
	//for scanner.Scan() {
	//	line := scanner.Text()
	//	_ = strings.Split(line, " ")
	//}

	// genericize
	tree := value.NewMapNode()
	value.Insert(tree, []string{"a", "b", "c"}, value.NewValnode(value.NewText("d")))
	value.Insert(tree, []string{"a", "b", "z"}, value.NewValnode(value.NewText("e")))
	fmt.Println(value.Tree2dot(tree))
}
