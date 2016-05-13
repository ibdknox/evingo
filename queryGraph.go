package main

import (
	"encoding/json"
	"fmt"
	"github.com/witheve/evingo/value"
	"strconv"
)

type Fact struct {
	entity    string
	attribute string
	value     value.Value
}

func (n Fact) String() string {
	return "{e: \"" + n.entity + "\", a: \"" + n.attribute + "\", v: " + n.value.String() + "}"
}

type Entity struct {
	entity     string
	attributes map[string]value.Value
}

func NewEntity(entity string) *Entity {
	return &Entity{entity: entity, attributes: make(map[string]value.Value)}
}

func (entity Entity) String() string {
	var result = "Entity<" + entity.entity + ">{"
	for attr, val := range entity.attributes {
		result += attr + ": " + val.String() + ", "
	}
	if len(entity.attributes) != 0 {
		result = result[:len(result)-2]
	}
	return result + "}"
}

type TagMap map[string][]*Entity

func (tagMap TagMap) String() string {
	result := "{\n"
	for k, entities := range tagMap {
		result += "  " + string(k) + ": [\n" // + item.String() + ",\n"
		for i, item := range entities {
			result += "    " + strconv.Itoa(i) + ": " + item.String() + ",\n"
		}
		result = result[:len(result)-2] + "\n  ],\n"
	}
	return result[:len(result)-2] + "\n}"
}

type BindingNode struct {
	variable *VariableNode
	field    string
	source   SourceNode
}

type VariableNode struct {
	name     string
	bindings []*BindingNode
}

type SourceNode interface {
	Bindings() *[]*BindingNode
}

type ScanNode struct {
	bindings []*BindingNode
}

func (source ScanNode) Bindings() []*BindingNode {
	return source.bindings
}

type ExpressionNode struct {
	operator   string
	bindings   []*BindingNode
	projection []*VariableNode
	grouping   []*VariableNode // ix's must be monotonically ordered integers
}

func (source ExpressionNode) Bindings() []*BindingNode {
	return source.bindings
}

type NotNode struct {
	body *QueryNode
}

type MemberNode struct {
	body *QueryNode
	ix   uint
}

type UnionNode struct {
	members []*MemberNode
}

type ChooseNode struct {
	members []*MemberNode
}

type MutateNode struct {
	operator string // add, remove, update
}

type QueryNode struct {
	name        string
	variables   map[string]*VariableNode
	expressions map[string]*ExpressionNode
	scans       map[string]*ScanNode
	mutates     map[string]*MutateNode
	nots        map[string]*NotNode
	unions      map[string]*UnionNode
	chooses     map[string]*ChooseNode
}

func NewQuery() *QueryNode {
	return &QueryNode{
		name:        "",
		variables:   make(map[string]*VariableNode),
		expressions: make(map[string]*ExpressionNode),
		scans:       make(map[string]*ScanNode),
		nots:        make(map[string]*NotNode),
		unions:      make(map[string]*UnionNode),
		chooses:     make(map[string]*ChooseNode),
	}
}

func panicOnNotOk(ok bool, msg string) {
	panic(msg)
}

//------------------------------------------------------------------------------
// Fact Fns
//------------------------------------------------------------------------------

func ReadFactsFromJson(raw []byte) *[]Fact {
	fmt.Println("reading facts", raw)
	var parsed [][]interface{}
	err := json.Unmarshal(raw, &parsed)
	if err != nil {
		panic(err)
	}

	var facts []Fact
	for k, v := range parsed {
		var fact = &Fact{entity: v[0].(string), attribute: v[1].(string)}

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

	return &facts
}

func FactsToEntities(factsPtr *[]Fact) *[]Entity {
	var facts = *factsPtr
	var entityMap = make(map[string]*Entity)

	for _, fact := range facts {
		var entity, ok = entityMap[fact.entity]
		if !ok {
			entity = NewEntity(fact.entity)
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

//------------------------------------------------------------------------------
// Entity Fns
//------------------------------------------------------------------------------
func IndexEntitiesById(entities []*Entity) *map[string]*Entity {
	var entityMap = make(map[string]*Entity)
	for _, entity := range entities {
		entityMap[entity.entity] = entity
	}
	return &entityMap
}

func IndexEntitiesByTag(entities *[]Entity) *TagMap {
	var tagMap = make(TagMap)
	var untagged = make([]*Entity, 0)
	for _, entity := range *entities {
		var tagValue, ok = entity.attributes["tag"]
		if ok {
			var tag = tagValue.(*value.Text).Value()
			var tagged, ok = tagMap[tag]
			if !ok {
				tagged = make([]*Entity, 0)
				tagMap[tag] = tagged
			}
			tagMap[tag] = append(tagged, &entity)

		} else {
			untagged = append(untagged, &entity)
		}
	}
	tagMap["$$untagged"] = untagged
	return &tagMap
}

type EntityFilter func(*Entity) bool

func EntityAttributeEquals(attribute string, value value.Value) EntityFilter {
	return func(entity *Entity) bool {
		var attrVal, ok = entity.attributes[attribute]
		if value == nil && ok == false {
			return true
		}
		if !ok || value == nil {
			return false
		}
		return attrVal.Equals(value)
	}
}

func FilterEntities(filter EntityFilter, entities []*Entity) []*Entity {
	var matches []*Entity
	for _, entity := range entities {
		if filter(entity) {
			matches = append(matches, entity)
		}
	}
	return matches
}

func SomeEntity(filter EntityFilter, entities []*Entity) *Entity {
	for _, entity := range entities {
		if filter(entity) {
			return entity
		}
	}
	return nil
}

//------------------------------------------------------------------------------
// QueryGraph Fns
//------------------------------------------------------------------------------

func QueryFromEntity(root *Entity, tagMap *TagMap) *QueryNode {
	var query = NewQuery()
	var nameValue = root.attributes["name"]
	if nameValue != nil {
		query.name = nameValue.(*value.Text).Value()
	}
	var sourceEntities = make(map[string]*Entity)
	var sources = make(map[string]SourceNode)

	var queryValue = value.NewText(root.entity)
	var queryChildFilter = EntityAttributeEquals("query", queryValue)
	var variableEntities = IndexEntitiesById(FilterEntities(queryChildFilter, (*tagMap)["variable"]))
	var scanEntities = FilterEntities(queryChildFilter, (*tagMap)["scan"])
	var expressionEntities = FilterEntities(queryChildFilter, (*tagMap)["expression"])
	var mutateEntities = FilterEntities(queryChildFilter, (*tagMap)["mutate"])

	// Prebuild everything that's going to cross-link (variables, scans, expressions, mutates)
	for _, entity := range *variableEntities {
		query.variables[entity.entity] = &VariableNode{name: entity.attributes["name"].(*value.Text).Value()}
	}
	for _, entity := range scanEntities {
		query.scans[entity.entity] = &ScanNode{}
		sourceEntities[entity.entity] = entity
	}
	for _, entity := range expressionEntities {
		query.expressions[entity.entity] = &ExpressionNode{operator: entity.attributes["operator"].(*value.Text).Value()}
		sourceEntities[entity.entity] = entity
	}
	for _, entity := range mutateEntities {
		query.mutates[entity.entity] = &MutateNode{operator: entity.attributes["operator"].(*value.Text).Value()}
		sourceEntities[entity.entity] = entity
	}

	// Build the binding nodes and link them into their variables and sources
	for id, variableEntity := range *variableEntities {
		var variable, ok = query.variables[id]
		panicOnNotOk(ok, "Query '"+query.name+"' does not contain variable '"+id+"'")
		for _, bindingEntity := range FilterEntities(EntityAttributeEquals("variable", value.NewText(variableEntity.entity)), (*tagMap)["binding"]) {
			var binding = &BindingNode{field: bindingEntity.attributes["field"].(*value.Text).Value()}
			binding.variable = variable
			var sourceId = bindingEntity.attributes["source"].(*value.Text).Value()
			var source, ok = sources[sourceId]
			panicOnNotOk(ok, "Query '"+query.name+"' does not contain source '"+sourceId+"' for binding '"+bindingEntity.entity+"'")
			*source.Bindings() = append(*source.Bindings(), binding)
			binding.source = source
		}
	}

	// Link projection and grouping variables to expression nodes
	for id, expression := range query.expressions {
		var expressionValue = value.NewText(id)

		for _, projectionEntity := range FilterEntities(EntityAttributeEquals("expression", expressionValue), (*tagMap)["projection"]) {
			var variableId = projectionEntity.attributes["variable"].(*value.Text).Value()
			var variable, ok = query.variables[variableId]
			panicOnNotOk(ok, "Query '"+query.name+"' does not contain variable '"+variableId+"' for projection '"+projectionEntity.entity+"'")
			expression.projection = append(expression.projection, variable)
		}

		var groupings = make(map[int64]*VariableNode)
		for _, groupingEntity := range FilterEntities(EntityAttributeEquals("expression", expressionValue), (*tagMap)["grouping"]) {
			var ix = groupingEntity.attributes["ix"].(*value.Number).Value().IntPart()
			var variableId = groupingEntity.attributes["variable"].(*value.Text).Value()
			var variable, ok = query.variables[variableId]
			panicOnNotOk(ok, "Query '"+query.name+"' does not contain variable '"+variableId+"' for grouping '"+groupingEntity.entity+"'")
			groupings[ix] = variable
		}
		var sortedGroupings = make([]*VariableNode, len(groupings))
		for ix, variable := range groupings {
			sortedGroupings[ix] = variable
		}
		expression.grouping = sortedGroupings
	}

	return query
}

func TagMapToQueryGraph(tagMap *TagMap) *QueryNode {

	var root = SomeEntity(EntityAttributeEquals("parent", nil), (*tagMap)["query"])
	if root == nil {
		panic("Unable to find root query!")
	}
	fmt.Println("query root entity", root.String())
	return QueryFromEntity(root, tagMap)
}
