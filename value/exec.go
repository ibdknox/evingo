package value

type Operator int

const (
	OpInsert = 1 + iota
	OpRemove
	OpFlush
	OpClose
	OpError
)

type evaluator func(Operator, []Value)
type builder func(Node, evaluator) evaluator

func buildNot(k Node, next evaluator) evaluator {
	return nil
}

func buildSum(k Node, next evaluator) evaluator {
	return nil
}

func buildScan(k Node, next evaluator) evaluator {
	return nil
}

// consider how to deal with the indices here
var builders = map[string]builder{
	"not":  buildNot,
	"sum":  buildSum,
	"scan": buildScan,
}

func build(source Node, final evaluator) {
}
