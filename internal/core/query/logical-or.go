package query

type logicalOr struct {
	items []Query
}

func (o logicalOr) Eval(p Params) bool {
	for _, item := range o.items {
		if item.Eval(p) {
			return true
		}
	}
	return false
}

func (o logicalOr) String() string {
	return sprintf("(%s)", join(o.items, " || "))
}

func Or(items ...Query) Query {
	return logicalOr{items}
}
