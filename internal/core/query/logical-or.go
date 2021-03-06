package query

type logicalOr struct {
	items []Query
}

func (self logicalOr) Eval(context Context) bool {
	for _, item := range self.items {
		if item.Eval(context) {
			return true
		}
	}
	return false
}

func (self logicalOr) String() string {
	return sprintf("(%s)", join(self.items, " || "))
}

func Or(items ...Query) Query {
	return &logicalOr{items}
}
