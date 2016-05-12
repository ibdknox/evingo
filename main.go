package main

import (
	//"bufio"
	"fmt"
	"github.com/witheve/evingo/parser"
	"github.com/witheve/evingo/util/color"
	"github.com/witheve/evingo/value"
	"os"
	//"strings"
)

func doDotStuff() {
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

func main() {
	args := os.Args
	argsLen := len(args)
	switch {
	case argsLen == 1:
		//print help
		fmt.Printf("Welcome to %s", color.Bright("Eve!\n"))
		fmt.Println("Here are the available commands:")
		fmt.Printf("  - %s\n", color.Bright("dot"))
		fmt.Printf("  - %s %s\n", color.Bright("parse"), color.Info("<file>"))
	case args[1] == "dot":
		doDotStuff()
	case args[1] == "parse":
		if argsLen > 2 {
			parser.ParseFile(args[2])
		} else {
			fmt.Println(color.Error("Must provide a file to parse"))
		}
	}

}
