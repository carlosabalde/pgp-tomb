package query

type logicalNot struct {
	item Query
}

func (n logicalNot) Eval(p Params) bool {
	return !n.item.Eval(p)
}

func (n logicalNot) String() string {
	return sprintf("!%s", n.item)
}

func Not(item Query) Query {
	return logicalNot{item}
}
