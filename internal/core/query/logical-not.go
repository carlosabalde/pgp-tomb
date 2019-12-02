package query

type logicalNot struct {
	item Query
}

func (self logicalNot) Eval(identifiers Identifiers) bool {
	return !self.item.Eval(identifiers)
}

func (self logicalNot) String() string {
	return sprintf("!%s", self.item)
}

func Not(item Query) Query {
	return logicalNot{item}
}
