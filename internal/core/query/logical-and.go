package query

type logicalAnd struct {
	items []Query
}

func (self logicalAnd) Eval(identifiers Identifiers) bool {
	for _, item := range self.items {
		if !item.Eval(identifiers) {
			return false
		}
	}
	return true
}

func (self logicalAnd) String() string {
	return sprintf("(%s)", join(self.items, " && "))
}

func And(items ...Query) Query {
	return logicalAnd{items}
}
