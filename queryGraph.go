package main

import (
	"encoding/json"
	"fmt"
	"github.com/witheve/evingo/value"
	"reflect"
	"strconv"
	"strings"
)

//------------------------------------------------------------------------------
// Utility Fns
//------------------------------------------------------------------------------

func panicOnNotOk(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func SliceToInterfaces(slice interface{}) []interface{} {
	var s = reflect.ValueOf(slice)
	panicOnNotOk(s.Kind() == reflect.Slice, "SliceToInterfaces() given a non-slice type")
	var ret = make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret
}

func MapToInterfaces(raw interface{}) map[string]interface{} {
	var m = reflect.ValueOf(raw)
	panicOnNotOk(m.Kind() == reflect.Map, "MapToInterfaces() given a non-map type")
	var ret = make(map[string]interface{})
	var keys = m.MapKeys()
	for _, key := range keys {
		var k = key.String()
		ret[k] = m.MapIndex(key).Interface()
	}
	return ret
}

func StringFromList(raw interface{}, indent int) string {
	var pad = ""
	for i := 0; i < indent; i++ {
		pad += "  "
	}
	var coll = SliceToInterfaces(raw)
	var result = "["
	for _, item := range coll {
		var stringer, ok = item.(fmt.Stringer)
		panicOnNotOk(ok, "Unable to coerce item to Stringer")
		var val = stringer.String()
		result += "\n  " + pad + strings.Join(strings.Split(val, "\n"), "\n  "+pad) + ","
	}
	if len(coll) > 0 {
		result = result[:len(result)-1]
	}
	return result + "\n" + pad + "]"
}

func StringFromMap(raw interface{}, indent int) string {
	var pad = ""
	for i := 0; i < indent; i++ {
		pad += "  "
	}

	var coll = MapToInterfaces(raw)
	var result = "{"
	for key, item := range coll {
		var val, ok = item.(fmt.Stringer)
		panicOnNotOk(ok, "Unable to coerce item for key '"+key+"' to Stringer")
		result += "\n  " + pad + key + ": " + strings.Join(strings.Split(val.String(), "\n"), "\n  "+pad) + ","
	}
	if len(coll) > 0 {
		result = result[:len(result)-1]
	}
	return result + "\n" + pad + "}"
}

func GetId(node fmt.Stringer) string {
	switch val := node.(type) {
	case *Entity:
		return val.entity
	case *BindingNode:
		return val.id
	case *VariableNode:
		return val.id
	case *ScanNode:
		return val.id
	case *ExpressionNode:
		return val.id
	case *MutateNode:
		return val.id
	case *NotNode:
		return val.id
	case *UnionNode:
		return val.id
	case *ChooseNode:
		return val.id
	case *QueryNode:
		return val.id
	}
	panic("Unknown node type, unable to fetch id: " + node.String())
	return ""
}

func StringFromIdList(coll interface{}) string {
	var result = "["
	var slice = SliceToInterfaces(coll)
	for _, item := range slice {
		var stringer, ok = item.(fmt.Stringer)
		panicOnNotOk(ok, "Unable to coerce item to Stringer")
		var id = GetId(stringer)
		result += "\"" + id + "\", "
	}
	if len(slice) > 0 {
		result = result[:len(result)-2]
	}
	return result + "]"
}

//------------------------------------------------------------------------------
// Types and Methods
//------------------------------------------------------------------------------

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
	id       string
	variable *VariableNode
	field    string
	source   SourceNode
}

func (binding *BindingNode) String() string {
	var result = "Binding<" + binding.id + ">{"
	result += "variable: " + binding.variable.id + ", "
	result += "field: " + binding.field + ", "
	result += "source: " + GetId(binding.source)
	return result + "}"
}

type VariableNode struct {
	id       string
	name     string
	bindings []*BindingNode
}

func (variable *VariableNode) String() string {
	var result = "Variable<" + variable.name + ">{"
	result += "\n  name: " + variable.name + ","
	result += "\n  bindings: (" + strconv.Itoa(len(variable.bindings)) + ") " + StringFromIdList(variable.bindings)
	return result + "\n}"
}

type SourceNode interface {
	Bindings() *[]*BindingNode
	String() string
}

type ScanNode struct {
	id       string
	bindings []*BindingNode
}

func (source *ScanNode) Bindings() *[]*BindingNode {
	return &source.bindings
}

func (source *ScanNode) String() string {
	return "Scan<" + source.id + ">{bindings: (" + strconv.Itoa(len(source.bindings)) + ") " + StringFromIdList(source.bindings) + "}"
}

type ExpressionNode struct {
	id         string
	operator   string
	bindings   []*BindingNode
	projection []*VariableNode
	grouping   []*VariableNode // ix's must be monotonically ordered integers
}

func (source *ExpressionNode) Bindings() *[]*BindingNode {
	return &source.bindings
}

func (source *ExpressionNode) String() string {
	var result = "Expression<" + source.id + ">{"
	result += "\n  operator: " + source.operator + ","
	result += "\n  bindings: (" + strconv.Itoa(len(source.bindings)) + ") " + StringFromIdList(source.bindings) + ","
	result += "\n  projection: (" + strconv.Itoa(len(source.projection)) + ") " + StringFromIdList(source.projection) + ","
	result += "\n  grouping: (" + strconv.Itoa(len(source.grouping)) + ") " + StringFromIdList(source.projection)
	return result + "\n}"
}

type MutateNode struct {
	id       string
	operator string // add, remove, update
	bindings []*BindingNode
}

func (source MutateNode) Bindings() *[]*BindingNode {
	return &source.bindings
}

func (source *MutateNode) String() string {
	return "Mutate<" + source.id + ">{operator: " + source.operator + "}"
}

type NotNode struct {
	id   string
	body *QueryNode
}

func (not *NotNode) String() string {
	return "Not<" + not.id + ">{body: " + not.body.id + "}"
}

type UnionNode struct {
	id      string
	members []*QueryNode
}

func (union *UnionNode) String() string {
	return "Union<" + union.id + ">{members: " + StringFromIdList(union.members) + "}"
}

type ChooseNode struct {
	id      string
	members []*QueryNode
}

func (choose *ChooseNode) String() string {
	return "Choose<" + choose.id + ">{members: " + StringFromIdList(choose.members) + "}"
}

type QueryNode struct {
	id          string
	name        string
	variables   map[string]*VariableNode
	expressions map[string]*ExpressionNode
	scans       map[string]*ScanNode
	mutates     map[string]*MutateNode
	nots        map[string]*NotNode
	unions      map[string]*UnionNode
	chooses     map[string]*ChooseNode
}

func (query QueryNode) String() string {
	var result = "Query<" + query.id + ">{"
	result += "\n  name: " + query.name + ","
	result += "\n  variables: " + StringFromMap(query.variables, 1) + ","
	result += "\n  expressions: " + StringFromMap(query.expressions, 1) + ","
	result += "\n  scans: " + StringFromMap(query.scans, 1) + ","
	result += "\n  mutates: " + StringFromMap(query.mutates, 1) + ","
	result += "\n  nots: " + StringFromMap(query.nots, 1) + ","
	result += "\n  unions: " + StringFromMap(query.unions, 1) + ","
	result += "\n  chooses: " + StringFromMap(query.chooses, 1)
	return result + "\n}"
}

func NewQuery(id string) *QueryNode {
	return &QueryNode{
		id:          id,
		name:        "",
		variables:   make(map[string]*VariableNode),
		expressions: make(map[string]*ExpressionNode),
		scans:       make(map[string]*ScanNode),
		nots:        make(map[string]*NotNode),
		unions:      make(map[string]*UnionNode),
		chooses:     make(map[string]*ChooseNode),
	}
}

//------------------------------------------------------------------------------
// Fact Fns
//------------------------------------------------------------------------------

func ReadFactsFromJson(raw []byte) *[]Fact {
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

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//
//
// @FIXME: attributes can't be collapsing like this. make everything a slice? :'(
//
//
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

func FactsToEntities(factsPtr *[]Fact) []*Entity {
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

	var entities = make([]*Entity, len(entityMap))
	var ix = 0
	for _, entity := range entityMap {
		entities[ix] = entity
		ix += 1
	}
	return entities
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

func IndexEntitiesByTag(entities []*Entity) *TagMap {
	var tagMap = make(TagMap)
	var untagged = make([]*Entity, 0)
	for _, entity := range entities {
		var tagValue, ok = entity.attributes["tag"]
		if ok {
			var tag = tagValue.(*value.Text).Value()
			var tagged, ok = tagMap[tag]
			if !ok {
				tagged = make([]*Entity, 0)
				tagMap[tag] = tagged
			}
			tagMap[tag] = append(tagged, entity)

		} else {
			untagged = append(untagged, entity)
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
	var query = NewQuery(root.entity)
	var nameValue = root.attributes["name"]
	if nameValue != nil {
		query.name = nameValue.(*value.Text).Value()
	}
	var sourceEntities = make(map[string]*Entity)
	var sources = make(map[string]SourceNode)

	var queryValue = value.NewText(query.id)
	var queryChildFilter = EntityAttributeEquals("query", queryValue)
	var variableEntities = IndexEntitiesById(FilterEntities(queryChildFilter, (*tagMap)["variable"]))
	var scanEntities = FilterEntities(queryChildFilter, (*tagMap)["scan"])
	var expressionEntities = FilterEntities(queryChildFilter, (*tagMap)["expression"])
	var mutateEntities = FilterEntities(queryChildFilter, (*tagMap)["mutate"])

	// Prebuild everything that's going to cross-link (variables, scans, expressions, mutates)
	for _, entity := range *variableEntities {
		query.variables[entity.entity] = &VariableNode{id: entity.entity, name: entity.attributes["name"].(*value.Text).Value()}
	}
	for _, entity := range scanEntities {
		query.scans[entity.entity] = &ScanNode{id: entity.entity}
		sourceEntities[entity.entity] = entity
		sources[entity.entity] = query.scans[entity.entity]
	}
	for _, entity := range expressionEntities {
		query.expressions[entity.entity] = &ExpressionNode{id: entity.entity, operator: entity.attributes["operator"].(*value.Text).Value()}
		sourceEntities[entity.entity] = entity
		sources[entity.entity] = query.expressions[entity.entity]
	}
	for _, entity := range mutateEntities {
		query.mutates[entity.entity] = &MutateNode{id: entity.entity, operator: entity.attributes["operator"].(*value.Text).Value()}
		sourceEntities[entity.entity] = entity
		sources[entity.entity] = query.mutates[entity.entity]
	}

	// Build the binding nodes and link them into their variables and sources
	for _, variableEntity := range *variableEntities {
		var variable, ok = query.variables[variableEntity.entity]
		panicOnNotOk(ok, "Query '"+query.name+"' does not contain variable '"+variableEntity.entity+"'")
		for _, bindingEntity := range FilterEntities(EntityAttributeEquals("variable", value.NewText(variable.id)), (*tagMap)["binding"]) {
			var binding = &BindingNode{id: bindingEntity.entity, field: bindingEntity.attributes["field"].(*value.Text).Value()}
			binding.variable = variable
			var sourceId = bindingEntity.attributes["source"].(*value.Text).Value()
			var source, ok = sources[sourceId]
			panicOnNotOk(ok, "Query '"+query.name+"' does not contain source '"+sourceId+"' for binding '"+bindingEntity.entity+"'")
			binding.source = source
			var bindings = source.Bindings()
			*bindings = append(*bindings, binding)
			variable.bindings = append(variable.bindings, binding)
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
