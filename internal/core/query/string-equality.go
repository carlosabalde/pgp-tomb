package query

type stringEquality struct {
	identifier string
	str        string
}

func (self stringEquality) Eval(context Context) bool {
	return context.GetIdentifier(self.identifier) == self.str
}

func (self stringEquality) String() string {
	return sprintf("(%s == '%s')", self.identifier, self.str)
}

func Equal(identifier, str string) Query {
	return &stringEquality{identifier, str}
}
