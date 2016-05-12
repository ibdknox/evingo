package main

import (
	"encoding/json"
	"fmt"
	"github.com/witheve/evingo/value"
)

type BindingNode struct {
	variable VariableNode
	field    string
	source   SourceNode
}

type VariableNode struct {
	name     string
	bindings []BindingNode
}

type SourceNode interface {
	Bindings() []BindingNode
}

type ScanNode struct {
	bindings []BindingNode
}

func (source ScanNode) Bindings() []BindingNode {
	return source.bindings
}

type ExpressionNode struct {
	op         string
	bindings   []BindingNode
	projection []VariableNode
	grouping   []VariableNode
}

func (source ExpressionNode) Bindings() []BindingNode {
	return source.bindings
}

type NotNode struct {
	body QueryNode
}

type MemberNode struct {
	body QueryNode
	ix   uint
}

type UnionNode struct {
	members []MemberNode
}

type ChooseNode struct {
	members []MemberNode
}

type MutationOperator uint8

const (
	add MutationOperator = iota
	remove
	update
)

type MutateNode struct {
	operator MutationOperator
}

type QueryNode struct {
	name        string
	variables   []VariableNode
	expressions []ExpressionNode
	scans       []ScanNode
	nots        []NotNode
	unions      []UnionNode
	chooses     []ChooseNode
}

type FactNode struct {
	entity    string
	attribute string
	value     value.Value
}

func (n FactNode) String() string {
	return "{e: \"" + n.entity + "\", a: \"" + n.attribute + "\", v: " + n.value.String() + "}"
}

func ReadFactsFromJson(raw []byte) []FactNode {
	fmt.Println("reading facts", raw)
	var parsed [][]interface{}
	err := json.Unmarshal(raw, &parsed)
	if err != nil {
		panic(err)
	}

	var facts []FactNode
	for k, v := range parsed {
		var fact = &FactNode{entity: v[0].(string), attribute: v[1].(string)}

		switch val := v[2].(type) {
		case string:
			fact.value = value.NewText(val)
		case int64:
			fact.value = value.NewNumberFromInt(val)
		case float64:
			fact.value = value.NewNumberFromFloat(val)
		case bool:
			fact.value = value.NewBoolean(val)
		default:
			fmt.Println("Unknown node type:", k)
			panic("Unknown node type!")
		}
		facts = append(facts, *fact)
	}

	return facts
}

type Entity struct {
	entity     string
	attributes map[string]value.Value
}

func (entity Entity) String() string {
	var result = "Entity<" + entity.entity + ">{"
	for attr, val := range entity.attributes {
		result += attr + ": " + val.String() + ", "
	}
	return result[:len(result)-2] + "}"
}

func FactsToEntities(facts []FactNode) *[]Entity {
	var entityMap = make(map[string]*Entity)

	for _, fact := range facts {
		var entity, ok = entityMap[fact.entity]
		if !ok {
			entity = &Entity{entity: fact.entity}
			entityMap[fact.entity] = entity
		}
		entity.attributes[fact.attribute] = fact.value
	}

	var entities = make([]Entity, len(entityMap))
	var ix = 0
	for _, entity := range entityMap {
		entities[ix] = *entity
		ix += 1
	}
	return &entities
}

type TagMap map[string][]Entity

func GroupEntitiesByTag(entities *[]Entity) *TagMap {
	var tagMap = make(TagMap)
	var untagged = make([]Entity, 0)
	for _, entity := range *entities {
		var tagValue, ok = entity.attributes["tag"]
		if ok {
			var tag = tagValue.String()
			var tagged, ok = tagMap[tag]
			if !ok {
				tagged = make([]Entity, 1)
				tagMap[tag] = tagged
			}
			tagMap[tag] = append(tagged, entity)

		} else {
			untagged = append(untagged, entity)
		}
	}
	return &tagMap
}

type EntityFilter func(Entity) bool

func EntityAttributeEquals(attribute string, value *value.Value) EntityFilter {
	return func(entity Entity) bool {
		var attrVal, ok = entity.attributes[attribute]
		if value == nil && ok == false {
			return true
		}
		if !ok || value == nil {
			return false
		}
		return attrVal.Equals(*value)
	}
}

func FilterEntities(filter EntityFilter, entities []Entity) *[]Entity {
	var matches []Entity
	for _, entity := range entities {
		if filter(entity) {
			matches = append(matches, entity)
		}
	}
	return &matches
}

func SomeEntity(filter EntityFilter, entities []Entity) *Entity {
	for _, entity := range entities {
		if filter(entity) {
			return &entity
		}
	}
	return nil
}

func TagMapToQueryGraph(tagMap *TagMap) *QueryNode {
	var queryNode = &QueryNode{}
	var root = SomeEntity(EntityAttributeEquals("parent", nil), (*tagMap)["query"])
	fmt.Println("query root entity", root.String())
	queryNode.name = root.attributes["name"].String()
	return queryNode
}
