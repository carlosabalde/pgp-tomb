package query

type logicalAnd struct {
	items []Query
}

func (self logicalAnd) Eval(context Context) bool {
	for _, item := range self.items {
		if !item.Eval(context) {
			return false
		}
	}
	return true
}

func (self logicalAnd) String() string {
	return sprintf("(%s)", join(self.items, " && "))
}

func And(items ...Query) Query {
	return &logicalAnd{items}
}
