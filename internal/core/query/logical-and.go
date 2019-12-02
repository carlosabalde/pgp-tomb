package query

type logicalAnd struct {
	items []Query
}

func (a logicalAnd) Eval(p Params) bool {
	for _, item := range a.items {
		if !item.Eval(p) {
			return false
		}
	}
	return true
}

func (a logicalAnd) String() string {
	return sprintf("(%s)", join(a.items, " && "))
}

func And(items ...Query) Query {
	return logicalAnd{items}
}
