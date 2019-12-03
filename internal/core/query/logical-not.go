package query

type logicalNot struct {
	item Query
}

func (self logicalNot) Eval(context Context) bool {
	return !self.item.Eval(context)
}

func (self logicalNot) String() string {
	return sprintf("!%s", self.item)
}

func Not(item Query) Query {
	return &logicalNot{item}
}
