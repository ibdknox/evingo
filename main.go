package main

import (
	//"bufio"
	"fmt"
	"github.com/witheve/evingo/parser"
	"github.com/witheve/evingo/util/color"
	"github.com/witheve/evingo/value"
	"io/ioutil"
	"os"
	"strconv"
	//"strings"
)

func panicOnError(e error) {
	if e != nil {
		panic(e)
	}
}

type Stringer interface {
	String() string
}

func pprintCollection(list *interface{}) {
}

func doDotStuff() {
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
		fmt.Printf("  - %s %s\n", color.Bright("parse"), color.Info("<file>"))
	case args[1] == "dot":
		doDotStuff()
	case args[1] == "parse":
		if argsLen > 2 {
			parser.ParseFile(args[2])
		} else {
			fmt.Println(color.Error("Must provide a file to parse"))
		}
	case args[1] == "load":
		fmt.Println("Loading", args[2])
		if argsLen > 2 {
			data, err := ioutil.ReadFile(args[2])
			panicOnError(err)
			var facts = ReadFactsFromJson(data)
			if true {
				fmt.Println("---FACTS---")
				result := "[\n"
				for k, item := range *facts {
					result += "  " + strconv.Itoa(k) + ": " + item.String() + ",\n"
				}
				fmt.Println(result[:len(result)-2] + "\n]")
			}

			var entities = FactsToEntities(facts)
			if true {
				fmt.Println("---ENTITIES---")
				result := "[\n"
				for k, item := range *entities {
					result += "  " + strconv.Itoa(k) + ": " + item.String() + ",\n"
				}
				fmt.Println(result[:len(result)-2] + "\n]")
			}

			var tagMap = GroupEntitiesByTag(entities)
			if true {
				fmt.Println("---TAG MAP---")
				fmt.Println(tagMap.String())
			}

			var query = TagMapToQueryGraph(tagMap)
			if true {
				fmt.Println("---QUERY GRAPH---")
				fmt.Println(query)
			}

			//scanner := bufio.NewScanner(os.Stdin)
			//for scanner.Scan() {
			//	line := scanner.Text()
			//	_ = strings.Split(line, " ")
			//}

		} else {
			fmt.Println(color.Error("Must provide a file to load"))
		}
	}
}
